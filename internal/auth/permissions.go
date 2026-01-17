package auth

const (
	PermPM2ViewBasic      = "pm2.view.basic"
	PermPM2ViewCwd        = "pm2.view.cwd"
	PermPM2ViewFull       = "pm2.view.full"
	PermPM2ControlStart   = "pm2.control.start"
	PermPM2ControlStop    = "pm2.control.stop"
	PermPM2ControlRestart = "pm2.control.restart"
)

const (
	PermF2BViewStatus   = "f2b.view.status"
	PermF2BViewJail     = "f2b.view.jail"
	PermF2BControlUnban = "f2b.control.unban"
)

const (
	PermUserView        = "user.view"
	PermUserCreate      = "user.create"
	PermUserEdit        = "user.edit"
	PermUserDelete      = "user.delete"
	PermUserRolesAssign = "user.roles.assign"
)

const (
	PermAuthLogin  = "auth.login"
	PermAuthLogout = "auth.logout"
	PermAuthVerify = "auth.verify"
)
