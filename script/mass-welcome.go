package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sort"
	"time"
)

const (
	EnvDiscordToken   = "DISCORD_TOKEN"
	EnvTargetServer   = "TARGET_SERVER"
	EnvWelcomeChannel = "WELCOME_CHANNEL"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env is not loaded")
	}
	discordToken := os.Getenv(EnvDiscordToken)
	welcomeChannel := os.Getenv(EnvWelcomeChannel)
	targetServer := os.Getenv(EnvTargetServer)

	bot, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		panic(err)
	}

	bot.Identify.Intents = discordgo.IntentsAll

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	members, err := bot.GuildMembers(targetServer, "", 1000)
	if err != nil {
		log.Printf("failed to get guild members, err: %v", err)
		return
	}
	newMembers := make([]discordgo.Member, 0, len(members))
	for _, user := range members {
		if user.JoinedAt.Before(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) {
			continue
		}
		fmt.Printf("joined at: %v %v\n", user.JoinedAt, user.User.Username)
		newMembers = append(newMembers, *user)
	}
	sort.Slice(newMembers, func(i, j int) bool {
		return newMembers[i].JoinedAt.Before(newMembers[j].JoinedAt)
	})
	for _, m := range newMembers {
		sendMassWelcomeMessages(bot, welcomeChannel, m)
		sendWelcomeDM(bot, m.User.ID)
	}
}

func sendMassWelcomeMessages(s *discordgo.Session, welcomeChannel string, m discordgo.Member) {
	embedMes := discordgo.MessageEmbed{
		Title:       "Lumosへようこそ!!!",
		Description: fmt.Sprintf("<@%s> さんがLumosにやってきました:sparkles:", m.User.ID),
		Color:       0xF1C40F,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    m.User.Username,
			URL:     fmt.Sprintf("https://discordapp.com/users/%s", m.User.ID),
			IconURL: m.User.AvatarURL(""),
		},
	}
	_, err := s.ChannelMessageSendEmbed(welcomeChannel, &embedMes)
	if err != nil {
		log.Printf("failed to send message, err: %v", err)
	}
}

func sendWelcomeDM(s *discordgo.Session, userID string) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		log.Printf("failed to create dm channel, err: %v", err)
		return
	}
	_, err = s.ChannelMessageSend(channel.ID, "# Lumosへようこそ :rainbow: \n## :sparkles:このサーバーについて\nサーバ名：Lumos(ルーモス)\n招待リンク：https://discord.gg/MTaq747KsS\n\n## :pushpin:メンバーへのお願い\n- お互いのことを貶したり馬鹿にする言動は謹んでください。\n- 初学者の方が多く参加しています。質問対応はお手柔らかにお願いします。\n- 仲間の投稿にはぜひスタンプなどで積極的に反応してあげてください。\n\n## :seedling:各カテゴリーの概要(一般,常設,ボイチャ,等)\n### お知らせ https://discord.com/channels/894226019240800276/937139360992747532\n- 運営がLumosメンバーへの告知に使用するテキストチャンネルです。\n- メンバーがこのカテゴリー内でメッセージを送ることは基本ありません。\n### ボイチャ https://discord.com/channels/894226019240800276/899646658504192060\n- Lumosの活動に必要なボイス&テキストチャンネルがまとめられています。\n### 常設\n- Lumosの活動に必要なテキストチャンネルが常設カテゴリー配下にあります。\n- 特に「目標宣言・達成報告」は全メンバーが使用するチャンネルになります。\n- どんな些細な情報でも「情報共有」に送っていただけるとLumosの活性化につながります。\n### プロジェクト https://discord.com/channels/894226019240800276/1114173103384301628\n- メンバー2人以上で立ち上げられるLumos内の活動単位です。\n- 詳しくは「プロジェクトアイデア」参照してください。\n\n## :beginner: 各チャンネルの概要\n### 情報共有・質問チャンネル https://discord.com/channels/894226019240800276/896409316955938867 https://discord.com/channels/894226019240800276/968126032836182016\n- 良い記事URL等ありましたら情報共有で共有してください。\n- エラーや環境構築などで困ったことがあれば、質問チャンネルで質問してください。\n### 目標宣言場・達成報告広場 https://discord.com/channels/894226019240800276/897844795639205919\n- これからやることを宣言してください！\n- その日１日でやり遂げたこと、自身の成果物を教えてください。\n- 仲間の投稿を眺めて、ぜひ勉強のモチベーションに繋げてください！\n### ボイチャ https://discord.com/channels/894226019240800276/899646658504192060\n- 「わいわい作業」はみんなで軽く雑談しながら作業するボイチャです。\n- 「モクモク作業」はミュート状態で集中して作業するボイチャです。")
	if err != nil {
		log.Printf("failed to send message, err: %v", err)
	}
}
