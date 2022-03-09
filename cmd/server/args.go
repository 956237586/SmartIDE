package server

import (
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var i18nInstance = i18n.GetInstance()

const (
	Flags_Mode = "mode"

	Flags_ServerWorkspaceid = "serverworkspaceid"
	Flags_ServerToken       = "servertoken"
	Flags_ServerUsername    = "serverusername"
	Flags_ServerUserGUID    = "serveruserguid"
	Flags_ServerFeedback    = "serverfeedback"
)

// 验证server模式下，flag是否有录入
func Check(cmd *cobra.Command) (err error) {

	fflags := cmd.Flags()

	// 如果不是 server 模式不需要验证
	mode, _ := fflags.GetString(Flags_Mode)
	if strings.ToLower(mode) != "server" {
		return nil
	}

	// server workspace id 不能为空；
	err = checkFlagRequired(fflags, Flags_ServerWorkspaceid)
	if err != nil {
		return err
	}

	// token 不能为空；
	err = checkFlagRequired(fflags, Flags_ServerToken)
	if err != nil {
		return err
	}

	// username、user guid不能为空；
	err = checkFlagRequired(fflags, Flags_ServerUsername)
	if err != nil {
		return err
	}
	err = checkFlagRequired(fflags, Flags_ServerUserGUID)
	if err != nil {
		return err
	}

	// feedback 地址不能为空
	err = checkFlagRequired(fflags, Flags_ServerFeedback)
	if err != nil {
		return err
	}

	common.SmartIDELog.Info("Mode server params validation passed.")

	return nil
}

// 检查参数是否填写
func checkFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	flagValue, _ := fflags.GetString(flagName)
	if !fflags.Changed(flagName) || flagValue == "" {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}
