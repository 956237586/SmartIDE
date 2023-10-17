#!/bin/bash
###########################################################################
# SmartIDE - Dev Containers
# Copyright (C) 2023 leansoftX.com

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
###########################################################################

USER_UID=${LOCAL_USER_UID:-1000}
USER_GID=${LOCAL_USER_GID:-1000}
USER_PASS=${LOCAL_USER_PASSWORD:-"smartide123.@IDE"}
USERNAME=smartide

ENV_PERSISTENT_HOME=${PERSISTENT_HOME:-"0"}
ENV_PERSISTENT_HOME_DIR=${PERSISTENT_HOME_DIR:-"/data/root"}
persist_home_dir() {
    HOME_DIR="$1"
    PERSIST_DIR="$2"
    if [ "${ENV_PERSISTENT_HOME}" = "1" ]; then
        if [ ! -d "${PERSIST_DIR}" ]; then
            mkdir -p "${PERSIST_DIR}"
            cp -r "${HOME_DIR}/." "${PERSIST_DIR}"
            echo "init $PERSIST_DIR"
        fi
        if [ -e "${HOME_DIR}" ] && [ ! -L "${HOME_DIR}" ]; then
            rm -rf "${HOME_DIR}-bak" && mv "${HOME_DIR}" "${HOME_DIR}-bak"
            echo "backup $HOME_DIR to ${HOME_DIR}-bak"
        fi
        if [ ! -d "${HOME_DIR}" ]; then
            ln -sf "${PERSIST_DIR}" "${HOME_DIR}"
            echo "recover $PERSIST_DIR to $HOME_DIR"
        fi
    fi
    rm -rf ~/.config/JetBrains/IdeaIC*/.lock
    echo "clean idea lock files"
}

ENV_PROJECTOR_SERVER_TOKEN=${PROJECTOR_SERVER_TOKEN:-""}
ENV_PROJECTOR_SERVER_RO_TOKEN=${PROJECTOR_SERVER_RO_TOKEN:-"$ENV_PROJECTOR_SERVER_TOKEN"}
set_projector_server_token(){
    if [ x"${ENV_PROJECTOR_SERVER_TOKEN}" = x"" ]; then
        echo "keep projector token empty"
        return
    fi
    CONFIG_DIR_NAME=$(cat /projector/ide/product-info.json |grep dataDirectoryName |awk -F'"' '{print $4}')
    IDEA_VM_FILE_DIR="$HOME/.config/JetBrains/$CONFIG_DIR_NAME"
    IDEA_VM_FILE="$IDEA_VM_FILE_DIR/idea64.vmoptions"
    echo "IDEA_VM_FILE:$IDEA_VM_FILE"
    if [ ! -d "${IDEA_VM_FILE_DIR}" ]; then
        mkdir -p "${IDEA_VM_FILE_DIR}"
        echo "create vm file dir $IDEA_VM_FILE_DIR"
    fi
    if [ ! -f "${IDEA_VM_FILE}" ]; then
        touch "$IDEA_VM_FILE"
        echo "create vm file  $IDEA_VM_FILE"
    fi
    TOKEN_LINE1_PREFIX="-DORG_JETBRAINS_PROJECTOR_SERVER_HANDSHAKE_TOKEN="
    TOKEN_LINE2_PREFIX="-DORG_JETBRAINS_PROJECTOR_SERVER_RO_HANDSHAKE_TOKEN="
    TOKEN_LINE1="$TOKEN_LINE1_PREFIX$ENV_PROJECTOR_SERVER_TOKEN"
    TOKEN_LINE2="$TOKEN_LINE2_PREFIX$ENV_PROJECTOR_SERVER_RO_TOKEN"
    # replace existing lines
    sed -i "/^${TOKEN_LINE1_PREFIX}/s#.*#${TOKEN_LINE1}#" "${IDEA_VM_FILE}"
    sed -i "/^${TOKEN_LINE2_PREFIX}/s#.*#${TOKEN_LINE2}#" "${IDEA_VM_FILE}"
    # Check again, if the lines do not exist, add them
    grep -q "^${TOKEN_LINE1_PREFIX}" "${IDEA_VM_FILE}" || echo "${TOKEN_LINE1}" >> "${IDEA_VM_FILE}"
    grep -q "^${TOKEN_LINE2_PREFIX}" "${IDEA_VM_FILE}" || echo "${TOKEN_LINE2}" >> "${IDEA_VM_FILE}"
    echo "set projector server rw token:$ENV_PROJECTOR_SERVER_TOKEN"
    echo "set projector server ro token:$ENV_PROJECTOR_SERVER_RO_TOKEN"
}

echo "gosu_entrypoint_node.sh"
echo "Starting with USER_UID : $USER_UID"
echo "Starting with USER_GID : $USER_GID"
echo "Starting with USER_PASS : $USER_PASS"

# root运行容器，容器里面一样root运行
if [ $USER_UID == '0' ]; then

    echo "-----root------Starting"

    USERNAMEROOT=root

    chown -R $USERNAMEROOT:$USERNAMEROOT /home/project
    #chown -R $USERNAMEROOT:$USERNAMEROOT /home/opvscode

    #chmod +x /home/opvscode/server.sh
    #ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    export HOME=/root

    persist_home_dir $HOME "$ENV_PERSISTENT_HOME_DIR"
    echo "root:$USER_PASS" | chpasswd

    set_projector_server_token
    echo "-----------Starting sshd"
    /usr/sbin/sshd

    echo "-----------Starting ide"
    exec /home/smartide/run.sh "$@"
else

    #非root运行，通过传入环境变量创建自定义用户的uid,gid，否则默认uid,gid为1000
    echo "-----smartide------Starting"

    # 启动传UID=1000  不需要修改UID，GID值
    if [[ $USER_UID != 1000 ]]; then
        echo "-----smartide---usermod uid start---"$(date "+%Y-%m-%d %H:%M:%S")
        usermod -u $USER_UID $USERNAME
        find / -user 1000 -exec chown -h $USERNAME {} \;
        echo "-----smartide---usermod uid end---"$(date "+%Y-%m-%d %H:%M:%S")
    fi

    if [[ $USER_GID != 1000 ]]; then
        echo "-----smartide---usermod gid start---"$(date "+%Y-%m-%d %H:%M:%S")
        # groupmod -g $USER_GID $USERNAME
        groupmod -g $USER_GID --non-unique $USERNAME
        find / -group 1000 -exec chgrp -h $USERNAME {} \;
        echo "-----smartide---usermod gid end---"$(date "+%Y-%m-%d %H:%M:%S")
    fi

    export HOME=/home/$USERNAME
    persist_home_dir $HOME "$ENV_PERSISTENT_HOME_DIR"
    # chmod g+rw /home
    chown -R $USERNAME:$USERNAME /home/project
    chown -R $USERNAME:$USERNAME "$ENV_PERSISTENT_HOME_DIR"
    mkdir -p /home/$USERNAME/.ssh
    chown -R $USERNAME:$USERNAME /home/$USERNAME/.ssh

    #chown -R $USERNAME:$USERNAME /home/opvscode
    #chmod +x /home/opvscode/server.sh

    echo "root:$USER_PASS" | chpasswd
    echo "smartide:$USER_PASS" | chpasswd

    # cp -r /root/.nvm /home/$USERNAME
    #ln -sf /home/$USERNAME/.nvm/versions/node/v$NODE_VERSION/bin/node /home/opvscode

    set_projector_server_token
    echo "-----smartide------Starting sshd"
    # do not detach (-D), log to stderr (-e), passthrough other arguments
    #exec /usr/sbin/sshd -D -e "$@"
    /usr/sbin/sshd
    
    echo "-----smartide-----Starting gosu ide"
    exec gosu smartide /home/smartide/run.sh "$@"
fi
