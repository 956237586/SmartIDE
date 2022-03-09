/*
 * @Author: jason chen
 * @Date: 2021-11-08
 * @Description: sqlite data access layer
 */

package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
)

// 工作模式
type WorkingMode string

const (
	WorkingMode_Remote WorkingMode = "remote"
	WorkingMode_Local  WorkingMode = "local"
	WorkingMode_K8s    WorkingMode = "k8s"
	WorkingMode_Server WorkingMode = "server"
)

// 远程连接的类型
type RemoteAuthType string

const (
	RemoteAuthType_SSH      RemoteAuthType = "ssh"
	RemoteAuthType_Password RemoteAuthType = "password"
)

// git库的连接方式
type GitRepoAuthType string

const (
	GitRepoAuthType_SSH   GitRepoAuthType = "ssh"
	GitRepoAuthType_HTTPS GitRepoAuthType = "https"
	GitRepoAuthType_HTTP  GitRepoAuthType = "http"
)

type WorkspaceType string

const (
	WorkspaceType_Server GitRepoAuthType = "server"
	WorkspaceType_Local  GitRepoAuthType = "local"
)

type WorkspaceInfo struct {
	ID                   string
	Name                 string
	WorkingDirectoryPath string
	// 配置文件路径
	ConfigFilePath string
	// 临时docker-compose文件生成后的保存路径
	TempDockerComposeFilePath string
	// 模式，local 本地、remote 远程
	Mode WorkingMode
	// host信息
	Remote RemoteInfo
	// git 库的克隆地址
	GitCloneRepoUrl string
	// WebIDE中文件所在根目录名称 （git库名称 或者 当前目录名）
	projectDirctoryName string
	// git 库的认证方式
	GitRepoAuthType GitRepoAuthType

	// 指定的分支
	Branch string

	// 配置文件
	ConfigYaml config.SmartIdeConfig

	K8sInfo K8sInfo

	// 临时的docker-compose文件
	TempDockerCompose compose.DockerComposeYml

	// 链接的docker-compose文件
	LinkDockerCompose compose.DockerComposeYml

	// 扩展信息
	Extend WorkspaceExtend

	// 创建时间
	CreatedTime time.Time

	// 关联的服务端workspace
	ServerWorkSpace model.ServerWorkspace
}

//
func CreateWorkspaceInfoFromServer(serverWorkSpace model.ServerWorkspace) (WorkspaceInfo, error) {
	repoName := getRepoName(serverWorkSpace.GitRepoUrl)
	workspaceInfo := WorkspaceInfo{
		ID:                   serverWorkSpace.NO,
		Name:                 serverWorkSpace.Name,
		ConfigFilePath:       serverWorkSpace.ConfigFilePath,
		GitCloneRepoUrl:      serverWorkSpace.GitRepoUrl,
		Branch:               serverWorkSpace.Branch,
		Mode:                 WorkingMode_Server,
		CreatedTime:          serverWorkSpace.CreatedAt,
		WorkingDirectoryPath: filepath.Join("~", "project", repoName),
		//TempDockerComposeFilePath: filepath.Join("~", REMOTE_REPO_ROOT, repoName),
		Remote: RemoteInfo{
			Addr:     serverWorkSpace.Resource.IP,
			UserName: serverWorkSpace.Resource.SSHUserName,
			Password: serverWorkSpace.Resource.SSHPassword,
			SSHPort:  model.CONST_Container_SSHPort,
		},
	}
	workspaceInfo.TempDockerComposeFilePath = workspaceInfo.GetTempDockerComposeFilePath()
	workspaceInfo.ServerWorkSpace = serverWorkSpace
	if serverWorkSpace.Extend != "" {
		err := json.Unmarshal([]byte(serverWorkSpace.Extend), &workspaceInfo.Extend)
		if err != nil {
			return WorkspaceInfo{}, err
		}
	}
	if serverWorkSpace.LinkDockerCompose != "" {
		err := yaml.Unmarshal([]byte(serverWorkSpace.LinkDockerCompose), &workspaceInfo.LinkDockerCompose)
		if err != nil {
			return WorkspaceInfo{}, err
		}
	}
	if serverWorkSpace.TempDockerComposeContent != "" {
		err := yaml.Unmarshal([]byte(serverWorkSpace.TempDockerComposeContent), &workspaceInfo.TempDockerCompose)
		if err != nil {
			return WorkspaceInfo{}, err
		}
	}
	if serverWorkSpace.ConfigFileContent != "" {
		err := yaml.Unmarshal([]byte(serverWorkSpace.ConfigFileContent), &workspaceInfo.ConfigYaml)
		if err != nil {
			return WorkspaceInfo{}, err
		}
	}

	workspaceInfo.ServerWorkSpace = serverWorkSpace

	return workspaceInfo, nil
}

// 工作区数据为空
func (w WorkspaceInfo) IsNil() bool {

	return w.ID == "" || w.WorkingDirectoryPath == "" || w.ConfigFilePath == "" || w.Name == "" || w.Mode == "" // || w.ProjectName == "" len(w.Extend.Ports) == 0 ||
}

// 工作区数据不为空
func (w WorkspaceInfo) IsNotNil() bool {
	return !w.IsNil()
}

// 验证
func (w WorkspaceInfo) Valid() error {
	/* if w.GetProjectDirctoryName() == "" {
		return errors.New("[Workspace] 项目名不能为空")
	} */

	if w.Mode == "" {
		return errors.New(i18nInstance.Main.Err_workspace_mode_none)

	}

	if w.ConfigFilePath == "" {
		return errors.New(i18nInstance.Main.Err_workspace_config_filepath_none)

	}

	if w.WorkingDirectoryPath == "" {
		return errors.New(i18nInstance.Main.Err_workspace_workingdir_none)

	}

	/* if w.GitCloneRepoUrl != "" {
		if !common.CheckGitRemoteUrl(w.GitCloneRepoUrl) {
			msg := fmt.Sprintf(i18nInstance.Main.Err_workspace_giturl_valid, w.GitCloneRepoUrl)
			return errors.New(msg)

		}
	} */

	return nil
}

//
func (w WorkspaceInfo) GetProjectDirctoryName() string {
	if w.projectDirctoryName == "" {
		if w.Mode == WorkingMode_Remote { // 远程模式
			if w.GitCloneRepoUrl == "" { // 当前模式下，不可能git库为空
				common.SmartIDELog.Error(i18nInstance.Common.Err_sshremote_param_repourl_none)
			}

			w.projectDirctoryName = getRepoName(w.GitCloneRepoUrl)
		} else if w.Mode == WorkingMode_Server {
			if w.GitCloneRepoUrl == "" { // 当前模式下，不可能git库为空
				common.SmartIDELog.Error(i18nInstance.Common.Err_sshremote_param_repourl_none)
			}

			w.projectDirctoryName = getRepoName(w.GitCloneRepoUrl)
		} else { // 本地模式
			//
			if w.GitCloneRepoUrl == "" && w.WorkingDirectoryPath == "" {
				common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_property_urlandworkingdir_none)
			}

			if w.GitCloneRepoUrl == "" { // 从工作目录中获取
				fileInfo, err := os.Stat(w.WorkingDirectoryPath)
				common.CheckError(err)
				w.projectDirctoryName = fileInfo.Name()
			} else { // 从git url中获取
				w.projectDirctoryName = getRepoName(w.GitCloneRepoUrl)
			}
		}

	}

	return w.projectDirctoryName
}

// 从 volumes 中获取容器工作目录
func (c *WorkspaceInfo) GetContainerWorkingPathWithVolumes() string {
	projectPath := ""

	service := c.TempDockerCompose.Services[c.ConfigYaml.Workspace.DevContainer.ServiceName]
	for _, volume := range service.Volumes {
		if strings.Contains(volume, ":/home/project") {
			tmp := strings.ReplaceAll(volume, "\\'", "")
			index := strings.Index(tmp, ":/home/project")
			projectPath = tmp[index+1:]
			break
		}
	}

	return projectPath
}

// get repo name
func getRepoName(repoUrl string) string {

	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}

//
func getLocalGitRepoUrl() (gitRemmoteUrl, pathName string) {
	// current directory
	pwd, err := os.Getwd()
	common.CheckError(err)
	fileInfo, err := os.Stat(pwd)
	common.CheckError(err)
	pathName = fileInfo.Name()

	// git remote url
	gitRepo, err := git.PlainOpen(pwd)
	//common.CheckError(err)
	if err == nil {
		gitRemote, err := gitRepo.Remote("origin")
		if err == nil {
			//common.CheckError(err)
			gitRemmoteUrl = gitRemote.Config().URLs[0]
		}
	}
	return gitRemmoteUrl, pathName
}

// 改变配置文件
func (w *WorkspaceInfo) ChangeConfig(currentConfigContent, linkDockerComposeContent string) (hasChanged bool) {
	// 参数检查
	if currentConfigContent == "" {
		msg := fmt.Sprintf(i18nInstance.Common.Warn_param_is_null, "configContent")
		common.SmartIDELog.Error(msg)
	}

	// 如果临时compose文件为空，那么肯定是改变
	if w.TempDockerCompose.IsNil() {
		return true
	}

	// 改变
	hasChanged = false // 默认为false
	ogriginConfigYamlContent, err := w.ConfigYaml.ToYaml()
	common.CheckError(err)
	if strings.ReplaceAll(currentConfigContent, " ", "") != strings.ReplaceAll(ogriginConfigYamlContent, " ", "") {
		var configYaml config.SmartIdeConfig
		err := yaml.Unmarshal([]byte(currentConfigContent), &configYaml)
		w.ConfigYaml = configYaml
		common.CheckError(err)
		hasChanged = true

	}
	originLinkComposeYamlContent, err := w.LinkDockerCompose.ToYaml()
	common.CheckError(err)
	if strings.ReplaceAll(linkDockerComposeContent, " ", "") != strings.ReplaceAll(originLinkComposeYamlContent, " ", "") {
		var linkDockerCompose compose.DockerComposeYml
		err := yaml.Unmarshal([]byte(linkDockerComposeContent), &linkDockerCompose)
		w.LinkDockerCompose = linkDockerCompose
		common.CheckError(err)
		hasChanged = true

	}

	return hasChanged
}

// 把结构化对象转换为string
func (instance *WorkspaceExtend) ToJson() string {

	d, err := json.Marshal(&instance)
	common.CheckError(err)

	return string(d)
}

//
func (instance *WorkspaceExtend) IsNotNil() bool {
	return !instance.IsNil()
}

//
func (instance *WorkspaceExtend) IsNil() bool {
	return instance == nil || len(instance.Ports) <= 0
}

// 工作区扩展字段
type WorkspaceExtend struct {
	// 端口映射情况
	Ports []config.PortMapInfo `json:"Ports"`
}

// 远程主机信息
type RemoteInfo struct {
	ID int
	// dns 或者 ip
	Addr        string
	UserName    string
	AuthType    RemoteAuthType
	Password    string
	SSHPort     int
	CreatedTime time.Time
}

type K8sInfo struct {
	ID             int
	CreatedTime    time.Time
	Context        string
	Namespace      string
	DeploymentName string
	PVCName        string
}

//
func (r RemoteInfo) IsNil() bool {
	return r.ID <= 0 || r.Addr == "" || r.UserName == "" || r.AuthType == ""
}

//
func (w RemoteInfo) IsNotNil() bool {
	return !w.IsNil()
}
