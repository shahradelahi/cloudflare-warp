package model

type IdentityAccount struct {
	Created                  string `json:"created"`
	Updated                  string `json:"updated"`
	License                  string `json:"license"`
	PremiumData              int64  `json:"premium_data"`
	WarpPlus                 bool   `json:"warp_plus"`
	AccountType              string `json:"account_type"`
	ReferralRenewalCountdown int64  `json:"referral_renewal_countdown"`
	Role                     string `json:"role"`
	ID                       string `json:"id"`
	Quota                    int64  `json:"quota"`
	Usage                    int64  `json:"usage"`
	ReferralCount            int64  `json:"referral_count"`
	TTL                      string `json:"ttl"`
}

type License struct {
	License string `json:"license"`
}
