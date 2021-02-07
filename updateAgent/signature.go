package upgradeAgent

import (
	"crypto"
	"encoding/base64"
	"github.com/ChenLong-dev/gobase/mbase/malgo"
)

var gServerPublicKey string = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1oKgFZF2bJJGZ4dyCAS/
irY4p8lEbR2VTh4vvubl9cbBJCPRMI27OvGTcwHWNf1/58Es+ky1ONvIQV/W8XoD
/fQ5XCw69oeg0QCC6BQxlruhAFZBpv2pK2j78s1po23cwiqqGMelmsZMJoa5r6na
xpF5eL2BzzGru/iNnDTa62PJCWJpmoEiLi0q4VAVv31zE4+JQE0mpl/YDNbRiy/I
N2ijBcbp6QwsBIkDySb1kSIynzKA+BwgeXnboSNHiL00KSs525ZweLm9PJdNXzph
mYVjPiG6ojmePoYTSGXDnj7rlJJlrx9qlUba/RyeKmNxLvqW/tEEeE7RFRAsD9ZU
4QIDAQAB
-----END PUBLIC KEY-----`
var gDeveloperPublicKey string = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxu6SXlgs9MgY6NBAX4qQ
CWMtICo6LMPLO8eTGZ9uZrjnPrYq43a6OB8pKCT2LyWLBHYI41tsRMH/JYr2j9S/
5u8Bg2Z2DRtGiIPSw40h4/byQmknzVZ0+0DwMpiJ4c5iOe7ln+B+PTuJjSfqbkqM
/0LQ9usTcmNHCHDhSlr/qOvBeVKlqYkw40K+RNyqkbnhce96SGK9XsmPD32U6jAJ
rawjosYcasZKR2sFHO9z9LpakT+eUvCmr3iQW+CLaZMVXAx5vdZPIyUrUguNP5h3
KzL/yBreH1KkZXvut/Dsm8Ah/UcCY2xjs5FfDXPIOVt99lYYGBjK3oCQgdimdXvr
3wIDAQAB
-----END PUBLIC KEY-----`

var serverRsa *malgo.Rsa = malgo.NewRsa(gServerPublicKey, "")
var developerRsa *malgo.Rsa = malgo.NewRsa(gDeveloperPublicKey, "")

func VerifyWithServer(data []byte, sign string) bool {
	//bsSign, bsErr := hex.DecodeString(sign)
	bsSign, bsErr := base64.StdEncoding.DecodeString(sign)
	if bsErr != nil {
		return false
	}

	return serverRsa.Verify(data, bsSign, crypto.SHA256)
}
func VerifyWithDeveloper(data []byte, sign string) bool {
	bsSign, bsErr := base64.StdEncoding.DecodeString(sign)
	if bsErr != nil {
		return false
	}

	return developerRsa.Verify(data, bsSign, crypto.SHA256)
}