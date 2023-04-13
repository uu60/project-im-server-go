package server

func GetPrefixedMessage(sender *User, msg string, isPrivate bool) string {
	msg = /*"[" + sender.Addr + "]" + */ "[" + sender.Name + "] " + msg
	if isPrivate {
		msg = "(私密对话)" + msg
	}
	return msg
}
