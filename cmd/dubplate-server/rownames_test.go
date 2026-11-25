package main

// The names of the rows doctor prints from the packages that own them,
// spelled as literals so a renamed row fails these tests rather than
// moving with them.
const (
	membersDoctorRow       = "members"
	authDirDoctorRow       = "auth dir"
	backupLastSuccessRow   = "backup last success"
	backupWALSizeRow       = "backup wal size"
	backupFreeSpaceRow     = "backup free space"
	backupLocalCopiesRow   = "backup local copies"
	backupRemoteRow        = "backup remote"
	unidecodeDoctorRow     = "unidecode"
	clientIPDoctorRow      = "client ip"
	bindAddrsDoctorRow     = "bind addresses"
	boundaryProbeDoctorRow = "boundary probe"
	webDoctorRow           = "web"
	musicVolumeDoctorRow   = "music volume"
)
