package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io"
    "net/http"
)

type KinMessage struct {
    ai_id string
    message string
}

type KinChatBreak struct {
    ai_id string
    greeting string
}

type KinDiscordBot struct {
    ShareCode    string         `json:"share_code"`
    EnableFilter bool           `json:"enable_filter"`
    Conversation []KinConversation `json:"conversation"`
}

type KinConversation struct {
    username string
    text string
    timestamp string
}

func (kin *NomiKin) NewConversationItem(username string, text string, timestamp string) KinConversation {
    return KinConversation{
        username: username,
        text: text,
        timestamp: timestamp,
    }
}

func (kin *NomiKin) SendKindroidApiCall(endpoint string, method string, body interface{}, extraHeaders map[string]string) ([]byte, error) {
    headers := map[string]string{
        "Authorization": "Bearer " + kin.ApiKey,
        "Content-Type": "application/json",
    }
    
    if endpoint == UrlComponents["DiscordBot"][0] {
        if extraHeaders["X-Kindroid-Requester"] == "" {
            return nil, fmt.Errorf("Error: X-Kindroid-Requester header is required for discord-bot endpoint")
        }
        headers["X-Kindroid-Requester"] = extraHeaders["X-Kindroid-Requester"]
    }

    var jsonBody []byte
    var bodyReader io.Reader
    var err error

    if body != nil {
        jsonBody, err = json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("Error constructing body: %v: ", err)
        }
        // DEBUGGING
        fmt.Printf("JSON Body: %v\n", string(jsonBody))
        bodyReader = bytes.NewBuffer(jsonBody)
    } else {
        bodyReader = nil
    }

    req, err := http.NewRequest(method, endpoint, bodyReader)
    if err != nil {
       return nil, fmt.Errorf("Error reading HTTP request: %v", err)
    }

    req.Header.Set("Authorization", headers["Authorization"])
    req.Header.Set("Content-Type", headers["Content-Type"])
    if endpoint == UrlComponents["DiscordBot"][0] {
        req.Header.Set("X-Kindroid-Requester", headers["X-Kindroid-Requester"])
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Error sending HTTP request: %v", err)
    }

    defer resp.Body.Close()

    bodyResponse, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Error response from Kin API: %v", string(bodyResponse))
    }

    log.Printf("Received reply message from Kin %v: %v", kin.CompanionId, string(bodyResponse))
    return bodyResponse, nil
}

func (kin *NomiKin) SendKindroidMessage(message *string) (string, error) {
    log.Printf("Sending message to Kin %v: %v", kin.CompanionId, message)
    endpoint := UrlComponents["SendMessage"][0]
    body := KinMessage{
        ai_id: kin.CompanionId,
        message: *message,
    }
    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", body, nil)
    if err != nil {
        return "", fmt.Errorf("Error sending message to Kin: %v", err)
    }


    kinReply := string(bodyResponse)
    return kinReply, nil
}

func (kin *NomiKin) SendKindroidChatBreak(message *KinChatBreak) (string, error) {
    log.Printf("Sending chat break to Kin %v: %v", kin.CompanionId, message.greeting)
    endpoint := UrlComponents["ChatBreak"][0]
    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", message, nil)
    if err != nil {
        return "", fmt.Errorf("Error sending chat break to Kin: %v", err)
    }

    kinReply := string(bodyResponse)
    if kinReply == "" {
        kinReply = "Chat break successful"
    }
    return kinReply, nil
}

func (kin *NomiKin) SendKindroidDiscordBot(kinShareId *string, discordNsfwFilter *bool, requester *string, conversation []KinConversation) (string, error) {
    log.Printf("Sending message to Kin %v: %v messages", kin.CompanionId, len(conversation))
    endpoint := UrlComponents["DiscordBot"][0]
    extraHeaders := map[string]string{
        "X-Kindroid-Requester": *requester,
    }
    body := KinDiscordBot{
        ShareCode: *kinShareId,
        EnableFilter: *discordNsfwFilter,
        Conversation: conversation,
    }

    bodyResponse, err := kin.SendKindroidApiCall(endpoint, "POST", body, extraHeaders)
    if err != nil {
        return "", fmt.Errorf("Error sending message to Kin: %v", err)
    }

    kinReply := string(bodyResponse)
    return kinReply, nil
}
