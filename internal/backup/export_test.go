package backup

// RemoteStatus lets the external tests run the remote row's probe
// under a bound of their own, which [DoctorRows] fixes.
var RemoteStatus = remoteStatus
