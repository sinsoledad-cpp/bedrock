package bootstrap

import "bedrock/pkg/validate"

func InitValidate() {
	if err := validate.InitTrans("zh"); err != nil {
		panic(err)
	}
}
