package upgradeAgent

type PackageInfo struct {
	Ver 				string		`json:"ver" bson:"ver"`
	DownloadUrl			string		`json:"downloadUrl" bson:"downloadUrl"`
	Proxy 				string		`json:"proxy" bson:"proxy"`
	SHA256 				string 		`json:"sha256" bson:"sha256"`
	Size 				uint64		`json:"size" bson:"size"`
	UpgradeCommand		string		`json:"upgradeCommand,omitempty" bson:"upgradeCommand"`
	ServerSignature		string		`json:"serverSignature" bson:"serverSignature"`
	DeveloperSignature	string		`json:"developerSignature" bson:"developerSignature"`
}

type PollRequest interface {
	NewStatusReport(ver string, result error)	StatusReport
}
type StatusReport interface {
}
type Response struct {
	Code 		int			`json:"code"`
	Message 	string		`json:"message"`
}