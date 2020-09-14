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
	userIDlist []string
	deadUserMap map[string]bool
	Token string
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
	if strings.HasPrefix(m.Content, "=명령어") {
		s.ChannelMessageSend(m.ChannelID, howToUse)
	}
	// =끔 @태그 로 특정 유저 뮤트
	if strings.HasPrefix(m.Content, "=끔") {
		for _, user := range m.Mentions {
			member, _ := s.State.Member(m.GuildID, user.ID)
			if member != nil {
				s.GuildMemberMute(m.GuildID, user.ID, true)
			}
		}
	}
	// =켬 @태그 로 특정 유저 뮤트 해제
	if strings.HasPrefix(m.Content, "=켬") {
		for _, user := range m.Mentions {
			member, _ := s.State.Member(m.GuildID, user.ID)
			if member != nil {
				s.GuildMemberMute(m.GuildID, user.ID, false)
			}
		}
	}
	// =자동참가 로 현재 채널 인원 전체 참가
	if strings.HasPrefix(m.Content, "=자동참가") {
		g, _ := s.Guild(m.GuildID)
		voiceList := g.VoiceStates
		for _, item := range voiceList {
			fmt.Println(item.ChannelID)
			userIDlist = append(userIDlist, item.UserID)
		}
	}
	// 현재 참가 인원 출력
	if strings.HasPrefix(m.Content, "=현재인원") {
		for _, user := range userIDlist {
			member, _ := s.State.Member(m.GuildID, user)
			s.ChannelMessageSend(m.ChannelID, member.Mention())
		}
	}
	// =참가 @태그 로 인원 참가
	if strings.HasPrefix(m.Content, "=참가") {
		for _, user := range m.Mentions {
			userIDlist = append(userIDlist, user.ID)
		}
	}
	// =나감 @태그 로 인원 나감
	if strings.HasPrefix(m.Content, "=나감") {
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
			}
		}
	}
	// 죽은 유저 리스트 업데이트
	if strings.HasPrefix(m.Content, "=뒤짐") {
		for _, user := range m.Mentions {
			deadUserMap[user.ID] = true
		}
	}
	// 죽은 유저 리스트 초기화
	if strings.HasPrefix(m.Content, "=새게임") {
		deadUserMap = make(map[string]bool, 100)
	}
	// 죽은 유저 살리기
	if strings.HasPrefix(m.Content, "=살림") {
		for _, user := range m.Mentions {
			if deadUserMap[user.ID] {
				deadUserMap[user.ID] = false
			}
		}
	}
	// 등록되어있고 살아있는 유저 뮤트 해제
	if strings.HasPrefix(m.Content, "=다켬") {
		for _, userid := range userIDlist {
			if !deadUserMap[userid] {
				member, _ := s.State.Member(m.GuildID, userid)
				if member != nil {
					s.GuildMemberMute(m.GuildID, userid, false)
				}
			}
		}
	}
	// 등록된 모든 유저 뮤트
	if strings.HasPrefix(m.Content, "=다끔") {
		for _, userid := range userIDlist {
			member, _ := s.State.Member(m.GuildID, userid)
			if member != nil {
				s.GuildMemberMute(m.GuildID, userid, true)
			}
		}
	}
}