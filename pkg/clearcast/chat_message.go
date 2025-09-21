package clearcast

const ChatMessageRoleSystem = "system"
const ChatMessageRoleUser = "user"
const ChatMessageRoleAssistant = "assistant"
const ChatMessageRoleFunction = "function"
const ChatMessageRoleTool = "tool"

func SystemMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    ChatMessageRoleSystem,
		Content: content,
	}
}

func UserMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    ChatMessageRoleUser,
		Content: content,
	}
}

func AssistantMessage(content string) ChatMessage {
	return ChatMessage{
		Role:    ChatMessageRoleAssistant,
		Content: content,
	}
}

func ToolMessage(callID string, content string) ChatMessage {
	return ChatMessage{
		Role:    ChatMessageRoleTool,
		Content: content,
	}
}

func FunctionMessage(callID string, content string) ChatMessage {
	return ChatMessage{
		Role:    ChatMessageRoleFunction,
		Content: content,
	}
}

// ChatMessage is a message in a chat conversation.
type ChatMessage struct {
	Role    string
	Content string
}
