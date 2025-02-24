package NomiKin

import (
    "log"
)

var Version = "v0.2.0"

type NomiKin struct {
    ApiKey      string
    CompanionId string
}

type NomiMessage struct {
    Text string
}

type NomiSentMessageContainer struct {
    SentMessage NomiMessage
}

type NomiReplyMessageContainer struct {
    ReplyMessage NomiMessage
}

type KinMessage struct {
    ai_id string
    message string
}

type KinChatBreak struct {
    ai_id string
    greeting string
}

type KinDiscordBot struct {
    share_code string
    enable_filter bool
    conversation []KinConversation
}

type KinConversation struct {
    username string
    text string
    timestamp string
}

var UrlComponents map[string][]string

func (nomi *NomiKin) Init(companionType string) {
    if companionType == "KINDROID" {
        log.Println("Kin Init")
        UrlComponents = make(map[string][]string)
        UrlComponents["SendMessage"] = []string {"https://api.kindroid.ai/v1/send-message"}
        UrlComponents["ChatBreak"] = []string {"https://api.kindroid.ai/v1/chat-break"}
        UrlComponents["DiscordBot"] = []string {"https://api.kindroid.ai/v1/discord-bot"}
    } else if companionType == "NOMI" {
        log.Println("Nomi Init")
        UrlComponents = make(map[string][]string)
        UrlComponents["SendMessage"] = []string {"https://api.nomi.ai/v1/nomis", "chat"}
        UrlComponents["RoomCreate"] = []string {"https://api.nomi.ai/v1/rooms"}
        UrlComponents["RoomReply"] = []string {"https://api.nomi.ai/v1/rooms", "chat/request"}
        UrlComponents["RoomSend"] = []string {"https://api.nomi.ai/v1/rooms", "chat"}
    } else {
        log.Println("Unknown companion type")
    }
}

