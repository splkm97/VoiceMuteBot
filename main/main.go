package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	howToUse string
	userList map[string]UserList
	Token    string
)

func init() {
	dat, err := ioutil.ReadFile("usage.txt")
	if err != nil {
		panic(err)
	}
	howToUse = string(dat)
	dat, err = ioutil.ReadFile("token.txt")
	if err != nil {
		panic(err)
	}
	userList = make(map[string]UserList)
	flag.StringVar(&Token, "t", string(dat), "Bot Token")
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
