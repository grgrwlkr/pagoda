package routenames

// Messenger route names
const (
	// Workspace routes
	MessengerWorkspaceList         = "messenger.workspace.list"
	MessengerWorkspaceView         = "messenger.workspace.view"
	MessengerWorkspaceCreate       = "messenger.workspace.create"
	MessengerWorkspaceUpdate       = "messenger.workspace.update"
	MessengerWorkspaceDelete       = "messenger.workspace.delete"
	MessengerWorkspaceAddMember    = "messenger.workspace.member.add"
	MessengerWorkspaceRemoveMember = "messenger.workspace.member.remove"

	// Channel routes
	MessengerChannelList         = "messenger.channel.list"
	MessengerChannelView         = "messenger.channel.view"
	MessengerChannelCreate       = "messenger.channel.create"
	MessengerChannelUpdate       = "messenger.channel.update"
	MessengerChannelDelete       = "messenger.channel.delete"
	MessengerChannelAddMember    = "messenger.channel.member.add"
	MessengerChannelRemoveMember = "messenger.channel.member.remove"
	MessengerChannelMessages     = "messenger.channel.messages"

	// Message routes
	MessengerMessageCreate  = "messenger.message.create"
	MessengerMessageUpdate  = "messenger.message.update"
	MessengerMessageDelete  = "messenger.message.delete"
	MessengerMessageReplies = "messenger.message.replies"
	MessengerMessageReply   = "messenger.message.reply"

	// Direct Message routes
	MessengerDMList          = "messenger.dm.list"
	MessengerDMView          = "messenger.dm.view"
	MessengerDMCreate        = "messenger.dm.create"
	MessengerDMMessages      = "messenger.dm.messages"
	MessengerDMMessageCreate = "messenger.dm.message.create"

	// Reaction routes
	MessengerReactionAdd    = "messenger.reaction.add"
	MessengerReactionRemove = "messenger.reaction.remove"

	// Attachment routes
	MessengerAttachmentUpload = "messenger.attachment.upload"
	MessengerAttachmentView   = "messenger.attachment.view"
	MessengerAttachmentDelete = "messenger.attachment.delete"
)
