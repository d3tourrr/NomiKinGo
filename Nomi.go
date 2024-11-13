package NomiKin

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
)

var UrlComponents map[string][]string

type Nomi struct {
    Uuid string
    Name string
}

type Room struct {
    Name string
    Uuid string
    Nomis []Nomi
}

type RoomContainer struct {
    Rooms []Room
}

type Message struct {
    Text string
}

type SentMessageContainer struct {
    SentMessage Message
}

type ReplyMessageContainer struct {
    ReplyMessage Message
}

func (nomi *NomiKin) Init() {
    log.Println("Entered Init")
    UrlComponents = make(map[string][]string)
    UrlComponents["SendMessage"] = []string {"https://api.nomi.ai/v1/nomis", "chat"}
    UrlComponents["RoomCreate"] = []string {"https://api.nomi.ai/v1/rooms"}
    UrlComponents["RoomSend"] = []string {"https://api.nomi.ai/v1/rooms", "chat"}
    UrlComponents["RoomReply"] = []string {"https://api.nomi.ai/v1/rooms", "chat/request"}
}

func (nomi *NomiKin) ApiCall(endpoint string, method string, body interface{}) ([]byte, error) {
    method = strings.ToUpper(method)

    headers := map[string]string{
        "Authorization": nomi.ApiKey,
        "Content-Type": "application/json",
    }

    var jsonBody []byte
    var bodyReader io.Reader
    var err error

    if body != nil {
        jsonBody, err = json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("Error constructing body: %v: ", err)
        }
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

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Error making HTTP request: %v", err)
    }

    defer resp.Body.Close()

    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading HTTP response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        var errorResult map[string]interface{}
        if err := json.Unmarshal(responseBody, &errorResult); err != nil {
            return nil, fmt.Errorf("Error unmarshalling error response: %v\n%v", err, string(responseBody))
        }

        return nil, fmt.Errorf("Error response from Nomi API: %v", string(responseBody))
    }

    return responseBody, nil
}

func (nomi *NomiKin) RoomExists(roomName *string) (*Room, error) {
    log.Printf("Checking Nomi %v room %v", nomi.CompanionId, *roomName)
    roomUrl := UrlComponents["RoomCreate"][0]
    roomResult, err := nomi.ApiCall(roomUrl, "Get", nil)

    if err != nil {
        return nil, err
    }

    var rooms RoomContainer
    if err := json.Unmarshal([]byte(roomResult), &rooms); err != nil {
        log.Printf("Cannot unmarshal to room: %v", string(roomResult))
        return nil, err
    }

    for _, r := range rooms.Rooms {
        if r.Name == *roomName {
            return &r, nil
        }
    }

    return nil, nil
}

func (nomi *NomiKin) CreateNomiRoom(name *string, note *string, backchannelingEnabled *bool, nomiUuids []string) (*Room, error) {
    roomCheck, err := nomi.RoomExists(name)
    if err != nil {
        log.Printf("Error checking if room exists: %v", err)
    }

    log.Printf("Room exists: %v. Nomi %v will be added if not already included.", name, nomi.CompanionId)
    if roomCheck != nil {
        inRoom := false
        for _, n := range roomCheck.Nomis {
            if n.Uuid == nomi.CompanionId {
                inRoom = true
                break
            }
        }

        if !inRoom {
            log.Printf("Adding Nomi %v to room %v", nomi.CompanionId, roomCheck.Name)
            roomNomis := []string{nomi.CompanionId}
            for _, n := range roomCheck.Nomis {
                roomNomis = append(roomNomis, n.Uuid)
            }

            roomNomisJson, err := json.Marshal(roomNomis)
            if err != nil {
                log.Printf("Error converting room Nomis %v to JSON: %v", roomNomis, err)
            }

            bodyMap := map[string]interface{}{
                "name": *name,
                "note": *note,
                "backchannelingEnabled": backchannelingEnabled,
                "nomiUuids": roomNomisJson,
            }

            response, err := nomi.ApiCall(UrlComponents["RoomCreate"][0], "Put", bodyMap)
            if err != nil {
                return nil, err
            }
            var result map[string]interface{}
            if err := json.Unmarshal([]byte(response), &result); err != nil {
                if roomCreateName, ok := result["name"].(string); ok {
                    log.Printf("Created Nomi %v room: %v\n", nomi.CompanionId, roomCreateName)
                    return &Room {Name: roomCreateName, Uuid: result["uuid"].(string)}, nil
                } else {
                    log.Printf("Error trying to create room %v: %v", bodyMap["name"], err)
                }
            }

        } else {
            log.Printf("Nomi %v is already in room %v", nomi.CompanionId, roomCheck.Name)
        }

        return &Room {Name: *name, Uuid: roomCheck.Uuid}, nil
    } else {
        log.Printf("Creating room: %v", *name)
        bodyMap := map[string]interface{}{
            "name": *name,
            "note": *note,
            "backchannelingEnabled": backchannelingEnabled,
            "nomiUuids": nomiUuids,
        }

        response, err := nomi.ApiCall(UrlComponents["RoomCreate"][0], "Post", bodyMap)
        if err != nil {
            return nil, err
        }

        var result map[string]interface{}
        if err := json.Unmarshal([]byte(response), &result); err != nil {
            if roomCreateName, ok := result["name"].(string); ok {
                log.Printf("Created Nomi %v room: %v\n", nomi.CompanionId, roomCreateName)
                return &Room {Name: roomCreateName, Uuid: result["uuid"].(string)}, nil
            } else {
                log.Printf("Error trying to create room %v: %v", bodyMap["name"], err)
            }
        }
    }

    return nil, fmt.Errorf("Failed to return anything meaningful")
}

func (nomi *NomiKin) SendNomiRoomMessage(message *string, roomId *string) (string, error) {
    if len(*message) > 599 {
        log.Printf("Message too long: %d", len(*message))
        return fmt.Sprintf("Your message was `%d` characters long, but the maximum message length is 600. Please send a shorter message.", len(*message)), nil
    }

    bodyMap := map[string]string{
        "messageText": *message,
    }

    messageSendUrl := UrlComponents["RoomSend"][0] + "/" + *roomId + "/" + UrlComponents["RoomSend"][1]
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }

    var result SentMessageContainer
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        log.Printf("Error parsing sent message response:\n %v", result)
    } else {
        log.Printf("Sent message to room %s: %v\n", *roomId, result.SentMessage.Text)
        return fmt.Sprintf("Sent message to room %s: %v\n", *roomId, result.SentMessage.Text), nil
    }

    return "", err
}

func (nomi *NomiKin) RequestNomiRoomReply(roomId *string, nomiId *string) (string, error) {
    bodyMap := map[string]string{
        "nomiUuid": *nomiId,
    }

    messageSendUrl := UrlComponents["RoomReply"][0] + "/" + *roomId + "/" + UrlComponents["RoomReply"][1]
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }


    var result ReplyMessageContainer
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        log.Printf("Error requesting Nomi %v response: %v", nomi.CompanionId, err)
    } else {
        log.Printf("Received Message from Nomi %v to room %s: %v\n", nomi.CompanionId, *roomId, result.ReplyMessage.Text)
        return result.ReplyMessage.Text, nil
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}

func (nomi *NomiKin) SendNomiMessage (message *string) (string, error) {
    if len(*message) > 599 {
        log.Printf("Message too long: %d", len(*message))
        return fmt.Sprintf("Your message was `%d` characters long, but the maximum message length is 600. Please send a shorter message.", len(*message)), nil
    }

    bodyMap := map[string]string{
        "messageText": *message,
    }

    bodyJson, err := json.Marshal(bodyMap)
    log.Printf("Sending message to Nomi %v: %v", nomi.CompanionId, string(bodyJson))

    messageSendUrl := UrlComponents["SendMessage"][0] + "/" + nomi.CompanionId + "/" + UrlComponents["SendMessage"][1]
    response, err := nomi.ApiCall(messageSendUrl, "Post", bodyMap)
    if err != nil {
        log.Printf("Error from API call: %v", err.Error())
    }

    var result map[string]interface{}
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        return "", err
    } else {
        if replyMessage, ok := result["replyMessage"].(map[string]interface{}); ok {
            log.Printf("Received reply message from Nomi %v: %v\n", nomi.CompanionId, replyMessage)
            if textValue, ok := replyMessage["text"].(string); ok {
                return textValue, nil
            }
        }
    }

    return "", fmt.Errorf("Failed to return anything meaningful")
}
