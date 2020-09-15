package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	userIDmap   map[string]bool
	deadUserMap map[string]bool
	Token       string
)

func init() {
	userIDmap = make(map[string]bool, 100)
	deadUserMap = make(map[string]bool, 100)
	flag.StringVar(&Token, "t", "NzQ5NTM2Mzg4NDI0MTM4Nzgy.X0taKA.Eg8o1Swg4hfSui7tsXA8HUNOSwo", "Bot Token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(messageCreate)
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = dg.Close()
}

func voiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	member, _ := s.GuildMember(vs.GuildID, vs.UserID)
	state := ""
	if vs.ChannelID == "" {
		state = ": exit from voice channel"
		delete(userIDmap, vs.UserID)
	}
	if vs.ChannelID != "" && !userIDmap[vs.UserID] {
		state = ": enter to voice channel <chID:" + vs.ChannelID + ">"
		userIDmap[vs.UserID] = true
	}
	fmt.Println(member.Nick + "'s voice state is changed" + state)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	const howToUse = "```\n* =현재인원\n현재 참가 상태인 인원을 출력합니다.\n\n* =참가 @태그\n특정 사용자를 참가시킵니다.\n참가된 사용자의 음성은 =다끔/=다켬 명령어에 영향을 받습니다.\n\n* =나감 @태그\n특정 사용자를 참가 해제시킵니다.\n참가 해제된 사용자의 음성은 =다끔/=다켬 명령어에 영향을 받지 않습니다.\n\n* =사용종료\n모든 사용자를 참가 해제시킵니다.\n\n* =뒤짐 @태그\n특정 사용자를 뒤짐 상태로 바꿉니다.\n뒤짐 상태의 사용자는 =다켬 명령어의 영향을 받지 않습니다.\n=다끔 명령어에는 영향을 받습니다.\n\n* =살림 @태그\n뒤짐 상태의 특정 사용자를 살림 상태로 바꿉니다.\n살림 상태의 사용자는 =다켬/=다끔 명령어의 영향을 받습니다..\n\n* =새게임\n모든 사용자를 살림 상태로 바꿉니다.\n\n* =끔 @태그\n특정 사용자의 마이크를 뮤트합니다.\n\n* =켬 @태그\n특정 사용자의 마이크를 뮤트 해제합니다.\n\nmade by 2km```"

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "=명령어" {
		_, _ = s.ChannelMessageSend(m.ChannelID, howToUse)
	}
	// =끔 @태그 로 특정 유저 뮤트
	if strings.HasPrefix(m.Content, "=끔") {
		muteByMention(s, m)
	}
	// =켬 @태그 로 특정 유저 뮤트 해제
	if strings.HasPrefix(m.Content, "=켬") {
		unMuteByMention(s, m)
	}
	// 현재 참가 인원 출력
	if strings.HasPrefix(m.Content, "=현재인원") {
		sendCurParticipants(s, m)
	}
	if strings.HasPrefix(m.Content, "=사용종료") {
		finishUse()
	}
	// =나감 @태그 로 인원 나감
	if strings.HasPrefix(m.Content, "=나감") {
		unParticipateByMention(s, m)
	}
	// 죽은 유저 리스트 업데이트
	if strings.HasPrefix(m.Content, "=뒤짐") {
		deadByMention(s, m)
	}
	// 죽은 유저 리스트 초기화
	if strings.HasPrefix(m.Content, "=새게임") {
		restoreAllCorps()
	}
	// 죽은 유저 살리기
	if strings.HasPrefix(m.Content, "=살림") {
		restoreByMention(s, m)
	}
	// 등록되어있고 살아있는 유저 뮤트 해제
	if strings.HasPrefix(m.Content, "=다켬") {
		unMuteAll(s, m)
	}
	// 등록된 모든 유저 뮤트
	if strings.HasPrefix(m.Content, "=다끔") {
		muteAll(s, m)
	}
}

func muteAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	for userID := range userIDmap {
		member, _ := s.State.Member(m.GuildID, userID)
		if member != nil {
			_ = s.GuildMemberMute(m.GuildID, userID, true)
		}
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "모두 뮤트하였습니다.")
}

func unMuteAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	for userID := range userIDmap {
		if !deadUserMap[userID] {
			member, _ := s.State.Member(m.GuildID, userID)
			if member != nil {
				_ = s.GuildMemberMute(m.GuildID, userID, false)
			}
		}
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "모두 뮤트 해제하였습니다.")
}

func restoreByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		if deadUserMap[user.ID] {
			deadUserMap[user.ID] = false
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 살렸습니다.")
	}
}

func restoreAllCorps() {
	deadUserMap = make(map[string]bool, 100)
}

func deadByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		deadUserMap[user.ID] = true
		_, _ = s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 죽였습니다.")
	}
}

func unParticipateByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		id := user.ID
		if userIDmap[id] {
			delete(userIDmap, id)
			_, _ = s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 내보냈습니다.")
		}
	}
}

func finishUse() {
	userIDmap = nil
}

func sendCurParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	if userIDmap != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "현재인원은 다음과 같습니다.")
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, "현재 아무도 참가되지 않았습니다.")
	}
	for user := range userIDmap {
		member, _ := s.State.Member(m.GuildID, user)
		if member != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, member.Mention())
		}
	}
}

func muteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		member, _ := s.State.Member(m.GuildID, user.ID)
		if member != nil {
			_ = s.GuildMemberMute(m.GuildID, user.ID, true)
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 뮤트했습니다.")
	}
}

func unMuteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		member, err := s.State.Member(m.GuildID, user.ID)
		fmt.Println(err)
		if member != nil {
			_ = s.GuildMemberMute(m.GuildID, user.ID, false)
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 뮤트 해제했습니다.")
	}
}
