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

type UserList struct {
	userIDmap   map[string]bool
	deadUserMap map[string]bool
}

var (
	userList map[string]UserList
	Token    string
)

func init() {
	userList = make(map[string]UserList)
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

func voiceStateUpdate(_ *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	if userList[vs.GuildID].userIDmap == nil || userList[vs.GuildID].deadUserMap == nil {
		userList[vs.GuildID] = UserList{make(map[string]bool), make(map[string]bool)}
	}
	if vs.ChannelID == "" {
		delete(userList[vs.GuildID].userIDmap, vs.UserID)
	}
	if vs.ChannelID != "" && !userList[vs.GuildID].userIDmap[vs.UserID] {
		userList[vs.GuildID].userIDmap[vs.UserID] = true
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	const howToUse = "```\n* =현재인원\n현재 참가 상태인 인원을 출력합니다.\n\n* =다끔\n모든 참가된 사용자를 뮤트시킵니다.\n\n* =다켬\n뒤짐 상태의 사용자를 제외한 모든 사용자를 뮤트시킵니다.\n\n* =뒤짐 @태그\n특정 사용자를 뒤짐 상태로 바꿉니다.\n뒤짐 상태의 사용자는 =다켬 명령어의 영향을 받지 않습니다.\n=다끔 명령어에는 영향을 받습니다.\n\n* =살림 @태그\n뒤짐 상태의 특정 사용자를 살림 상태로 바꿉니다.\n살림 상태의 사용자는 =다켬/=다끔 명령어의 영향을 받습니다..\n\n* =새게임\n모든 사용자를 살림 상태로 바꿉니다.\n\n* =끔 @태그\n특정 사용자의 마이크를 뮤트합니다.\n\n* =켬 @태그\n특정 사용자의 마이크를 뮤트 해제합니다.\n\nmade by 2km```"

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
	// 죽은 유저 리스트 업데이트
	if strings.HasPrefix(m.Content, "=뒤짐") {
		deadByMention(s, m)
	}
	// 죽은 유저 리스트 초기화
	if strings.HasPrefix(m.Content, "=새게임") {
		restoreAllCorps(s, m)
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
	for userID := range userList[m.GuildID].userIDmap {
		go func(gid, uid string) {
			member, _ := s.GuildMember(gid, uid)
			if member != nil {
				_ = s.GuildMemberMute(gid, uid, true)
			}
		}(m.GuildID, userID)
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "모두 뮤트하였습니다.")
}

func unMuteAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	for userID := range userList[m.GuildID].userIDmap {
		go func(gid, uid string) {
			if !userList[gid].deadUserMap[uid] {
				member, _ := s.GuildMember(gid, uid)
				if member != nil {
					_ = s.GuildMemberMute(gid, uid, false)
				}
			}
		}(m.GuildID, userID)
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "모두 뮤트 해제하였습니다.")
}

func restoreByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		if userList[m.GuildID].deadUserMap[user.ID] {
			userList[m.GuildID].deadUserMap[user.ID] = false
		}
		go func(cid, msg string) {
			_, _ = s.ChannelMessageSend(cid, msg)
		}(m.ChannelID, user.Username+"을(를) 살렸습니다.")
	}
}

func restoreAllCorps(s *discordgo.Session, m *discordgo.MessageCreate) {
	for member := range userList[m.GuildID].deadUserMap {
		delete(userList[m.GuildID].deadUserMap, member)
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "모두 살렸습니다.")
}

func deadByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, user := range m.Mentions {
		go func(gid, cid, uid, msg string) {
			userList[gid].deadUserMap[uid] = true
			_, _ = s.ChannelMessageSend(cid, msg)
		}(m.GuildID, m.ChannelID, user.ID, user.Mention()+"을(를) 죽였습니다.")
	}
}

func sendCurParticipants(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := ""
	for user := range userList[m.GuildID].userIDmap {
		member, _ := s.GuildMember(m.GuildID, user)
		msg += "<" + member.User.Username + "> "
	}
	if len(userList[m.GuildID].userIDmap) != 0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "현재인원은 다음과 같습니다.")
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, "현재 아무도 참가되지 않았습니다.")
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, msg)
}

func muteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := ""
	for _, user := range m.Mentions {
		if user.ID != "" {
			go func(gid, uid string) {
				_ = s.GuildMemberMute(gid, uid, true)
			}(m.GuildID, user.ID)
			msg += user.Username + "을(를) 뮤트했습니다.\n"
		}
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, msg)
}

func unMuteByMention(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := ""
	for _, user := range m.Mentions {
		go func(gid, uid string) {
			member, _ := s.GuildMember(gid, uid)
			if member != nil {
				_ = s.GuildMemberMute(gid, uid, false)
			}
		}(m.GuildID, user.ID)
		msg += user.Username + "을(를) 뮤트 해제했습니다.\n"
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, msg)
}
