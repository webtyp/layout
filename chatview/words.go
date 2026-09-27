package chatview

import "webtyp.com/fmt/lang"

func init() {
	lang.RegisterWords([]lang.DictEntry{
		{EN: "Conversations", ES: "Conversaciones"},
		{EN: "People", ES: "Personas"},
		{EN: "Send", ES: "Enviar"},
		{EN: "Write a message", ES: "Escribe un mensaje"},
		{EN: "Read", ES: "Leído"},
		{EN: "Online", ES: "En línea"},
		{EN: "Offline", ES: "Desconectado"},
		{EN: "No conversations yet", ES: "Aún no hay conversaciones"},
		{EN: "No messages yet", ES: "Aún no hay mensajes"},
		{EN: "Pick a conversation", ES: "Elige una conversación"},
		{EN: "Nobody else is here yet", ES: "Todavía no hay nadie más"},
	})
}
