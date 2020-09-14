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
	userIDlist  []string
	deadUserMap map[string]bool
	Token       string
)

func init() {
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
	dg.Close()
}
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	const howToUse = "```\n* =자동참가\n현재 음성 채널에 있는 모든 사용자를 참가시킵니다.\n참가된 사용자의 음성은 =다끔/=다켬 명령어에 영향을 받습니다.\n\n* =참가 @태그\n특정 사용자를 참가시킵니다.\n참가된 사용자의 음성은 =다끔/=다켬 명령어에 영향을 받습니다.\n\n* =나감 @태그\n특정 사용자를 참가 해제시킵니다.\n참가 해제된 사용자의 음성은 =다끔/=다켬 명령어에 영향을 받지 않습니다.\n\n* =사용종료\n모든 사용자를 참가 해제시킵니다.\n\n* =뒤짐 @태그\n특정 사용자를 뒤짐 상태로 바꿉니다.\n뒤짐 상태의 사용자는 =다켬 명령어의 영향을 받지 않습니다.\n=다끔 명령어에는 영향을 받습니다.\n\n* =살림 @태그\n뒤짐 상태의 특정 사용자를 살림 상태로 바꿉니다.\n살림 상태의 사용자는 =다켬/=다끔 명령어의 영향을 받습니다..\n\n* =새게임\n모든 사용자를 살림 상태로 바꿉니다.\n\n* =끔 @태그\n특정 사용자의 마이크를 뮤트합니다.\n\n* =켬 @태그\n특정 사용자의 마이크를 뮤트 해제합니다.\n\nmade by 2km```"

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "=명령어" {
		s.ChannelMessageSend(m.ChannelID, howToUse)
	}
	// =끔 @태그 로 특정 유저 뮤트
	if strings.HasPrefix(m.Content, "=끔") {
		muteByMention(s, m)
	}
	// =켬 @태그 로 특정 유저 뮤트 해제
	if strings.HasPrefix(m.Content, "=켬") {
		unMuteByMention(s, m)
	}
	// =자동참가 로 현재 채널 인원 전체 참가
	if m.Content == "=자동참가" {
		autoParticipate(s, m)
	}
	// 현재 참가 인원 출력
	if strings.HasPrefix(m.Content, "=현재인원") {
		sendCurParticipants(s, m)
	}
	if strings.HasPrefix(m.Content, "=사용종료") {
		finishUse()
	}
	// =참가 @태그 로 인원 참가
	if strings.HasPrefix(m.Content, "=참가") {
		participateByMention(m)
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
	for _, userid := range userIDlist {
		member, _ := s.State.Member(m.GuildID, userid)
		if member != nil {
			s.GuildMemberMute(m.GuildID, userid, true)
		}
	}
	s.ChannelMessageSend(m.ChannelID, "모두 뮤트하였습니다.")
}

func unMuteAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, userid := range userIDlist {
		if !deadUserMap[userid] {
			member, _ := s.State.Member(m.GuildID, userid)
			if member != nil {
				s.GuildMemberMute(m.GuildID, userid, false)
			}
		}
	}
	s.ChannelMessageSend(m.ChannelID, "모두 뮤트 해제하였습니다.")
}

func restoreByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		if deadUserMap[user.ID] {
			deadUserMap[user.ID] = false
		}
		s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 살렸습니다.")
	}
}

func restoreAllCorps() {
	deadUserMap = make(map[string]bool, 100)
}

func deadByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		deadUserMap[user.ID] = true
		s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 죽였습니다.")
	}
}

func unParticipateByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		index := -1
		id := user.ID
		for i, listedUser := range userIDlist {
			if id == listedUser {
				index = i
			}
		}
		if index != -1 {
			userIDlist = append(userIDlist[:index], userIDlist[index+1:]...)
			s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 내보냈습니다.")
		}
	}
}

func participateByMention(m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		userIDlist = append(userIDlist, user.ID)
	}
}

func finishUse() {
	userIDlist = nil
}

func sendCurParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	if userIDlist != nil {
		s.ChannelMessageSend(m.ChannelID, "현재인원은 다음과 같습니다.")
	} else {
		s.ChannelMessageSend(m.ChannelID, "현재 아무도 참가되지 않았습니다.")
	}
	for _, user := range userIDlist {
		member, _ := s.State.Member(m.GuildID, user)
		if member != nil {
			s.ChannelMessageSend(m.ChannelID, member.Mention())
		}
	}
}

func autoParticipate(s *discordgo.Session, m *discordgo.MessageCreate) {
	g, _ := s.Guild(m.GuildID)
	vList := g.VoiceStates
	for _, vState := range vList {
		fmt.Println(vState.UserID)
		userIDlist = append(userIDlist, vState.UserID)
	}
}

func muteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		member, _ := s.State.Member(m.GuildID, user.ID)
		if member != nil {
			s.GuildMemberMute(m.GuildID, user.ID, true)
		}
		s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 뮤트했습니다.")
	}
}

func unMuteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		member, _ := s.State.Member(m.GuildID, user.ID)
		if member != nil {
			s.GuildMemberMute(m.GuildID, user.ID, false)
		}
		s.ChannelMessageSend(m.ChannelID, user.Mention()+"을(를) 뮤트 해제했습니다.")
	}
}
