package routenames

import (
	"fmt"
)

const (
	Home                 = "home"
	About                = "about"
	Contact              = "contact"
	ContactSubmit        = "contact.submit"
	Login                = "login"
	LoginSubmit          = "login.submit"
	Register             = "register"
	RegisterSubmit       = "register.submit"
	ForgotPassword       = "forgot_password"
	ForgotPasswordSubmit = "forgot_password.submit"
	Logout               = "logout"
	VerifyEmail          = "verify_email"
	ResetPassword        = "reset_password"
	ResetPasswordSubmit  = "reset_password.submit"
	Search               = "search"
	Task                 = "task"
	TaskSubmit           = "task.submit"
	Cache                = "cache"
	CacheSubmit          = "cache.submit"
	Files                = "files"
	FilesSubmit          = "files.submit"
	AdminTasks           = "admin:tasks"
	WebSocket            = "websocket"

	// Messenger route names
	// Root route
	MessengerRoot = "messenger.root"

	// Workspace routes
	MessengerWorkspaceList         = "messenger.workspace.list"
	MessengerWorkspaceView         = "messenger.workspace.view"
	MessengerWorkspaceCreate       = "messenger.workspace.create"
	MessengerWorkspaceCreateForm   = "messenger.workspace.create.form"
	MessengerWorkspaceUpdate       = "messenger.workspace.update"
	MessengerWorkspaceDelete       = "messenger.workspace.delete"
	MessengerWorkspaceAddMember    = "messenger.workspace.member.add"
	MessengerWorkspaceRemoveMember = "messenger.workspace.member.remove"

	// Channel routes
	MessengerChannelList         = "messenger.channel.list"
	MessengerChannelView         = "messenger.channel.view"
	MessengerChannelCreate       = "messenger.channel.create"
	MessengerChannelCreateForm   = "messenger.channel.create.form"
	MessengerChannelUpdate       = "messenger.channel.update"
	MessengerChannelDelete       = "messenger.channel.delete"
	MessengerChannelAddMember    = "messenger.channel.member.add"
	MessengerChannelRemoveMember = "messenger.channel.member.remove"
	MessengerChannelMessages     = "messenger.channel.messages"

	// Message routes
	MessengerMessageCreate      = "messenger.message.create"
	MessengerMessageUpdate      = "messenger.message.update"
	MessengerMessageDelete      = "messenger.message.delete"
	MessengerMessageReplies     = "messenger.message.replies"
	MessengerMessageReply       = "messenger.message.reply"
	MessengerMessageThreadPanel = "messenger.message.thread.panel"

	// Direct Message routes
	MessengerDirectMessageList          = "messenger.direct_message.list"
	MessengerDirectMessageView          = "messenger.direct_message.view"
	MessengerDirectMessageCreate        = "messenger.direct_message.create"
	MessengerDirectMessageMessages      = "messenger.direct_message.messages"
	MessengerDirectMessageMessageCreate = "messenger.direct_message.message.create"

	// Reaction routes
	MessengerReactionAdd    = "messenger.reaction.add"
	MessengerReactionRemove = "messenger.reaction.remove"

	// Attachment routes
	MessengerAttachmentUpload              = "messenger.attachment.upload"
	MessengerDirectMessageAttachmentUpload = "messenger.direct_message.attachment.upload"
	MessengerAttachmentView                = "messenger.attachment.view"
	MessengerAttachmentDelete              = "messenger.attachment.delete"

	// Search routes
	MessengerSearch         = "messenger.search"
	MessengerSearchMessages = "messenger.search.messages"
	MessengerSearchUsers    = "messenger.search.users"
)

func AdminEntityList(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_list", entityTypeName)
}

func AdminEntityAdd(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_add", entityTypeName)
}

func AdminEntityEdit(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_edit", entityTypeName)
}

func AdminEntityDelete(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_delete", entityTypeName)
}

func AdminEntityAddSubmit(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_add.submit", entityTypeName)
}

func AdminEntityEditSubmit(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_edit.submit", entityTypeName)
}

func AdminEntityDeleteSubmit(entityTypeName string) string {
	return fmt.Sprintf("admin:%s_delete.submit", entityTypeName)
}
