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

# FROM ubuntu:20.04 AS ideDownloader
# # -------------下载IDE文件
# # prepare tools:
# RUN apt-get update && apt-get -y install --no-install-recommends wget ca-certificates tar
# # download IDE to the /ide dir:
# WORKDIR /download

# # https://download.jetbrains.com/idea/ideaIC-2021.2.3.tar.gz
# ARG downloadUrl
# RUN wget -q $downloadUrl -O - | tar -xz
# RUN find . -maxdepth 1 -type d -name * -execdir mv {} /ide \;

# # -------------构建projector-server库，运行依赖
# FROM smartide/projector-server:latest as projectorGradleBuilder


# # -------------处理IDE运行程序，和构建好的projector-server库
# FROM ubuntu:20.04 AS projectorStaticFiles

# # prepare tools:
# RUN apt-get update && apt-get -y install --no-install-recommends unzip
# # create the Projector dir:
# ENV PROJECTOR_DIR /projector
# RUN mkdir -p $PROJECTOR_DIR
# # copy IDE:
# COPY --from=ideDownloader /ide $PROJECTOR_DIR/ide
# # copy projector files to the container:
# COPY static $PROJECTOR_DIR
# # copy projector:
# COPY --from=projectorGradleBuilder $PROJECTOR_DIR/projector-server/projector-server/build/distributions/projector-server.zip $PROJECTOR_DIR
# # prepare IDE - apply projector-server:
# RUN unzip $PROJECTOR_DIR/projector-server.zip
# RUN rm $PROJECTOR_DIR/projector-server.zip
# RUN find . -maxdepth 1 -type d -name projector-server-* -exec mv {} projector-server \;
# RUN mv projector-server $PROJECTOR_DIR/ide/projector-server
# RUN mv $PROJECTOR_DIR/ide-projector-launcher.sh $PROJECTOR_DIR/ide/bin
# RUN chmod 644 $PROJECTOR_DIR/ide/projector-server/lib/*



# -------------配置运行环境

# FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2:latest
FROM registry.cn-hangzhou.aliyuncs.com/hylstudio/smartide-java-v2:2022060-2b071e5ae5589
RUN true \
# Any command which returns non-zero exit code will cause this shell script to exit immediately:
   && set -e \
# Activate debugging to show execution details: all commands will be printed before execution
   && set -x \
# install packages:
    && apt-get update \
# packages for awt:
    && apt-get install --no-install-recommends libxext6 libxrender1 libxtst6 libxi6 libfreetype6 -y \
# packages for user convenience:
    && apt-get install --no-install-recommends vim git wget curl bash-completion net-tools sudo ca-certificates procps tar unzip tmux -y \
# clean apt to reduce image size:
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/cache/apt
# create the Projector dir:
ENV PROJECTOR_DIR /projector
RUN mkdir -p $PROJECTOR_DIR
# copy IDE:
# COPY --from=ideDownloader /ide $PROJECTOR_DIR/ide
# https://download.jetbrains.com/idea/ideaIC-2023.1.tar.gz
ARG downloadUrl
RUN mkdir /download && cd /download && wget -q $downloadUrl -O - | tar -xz && \
    cd /download && find . -maxdepth 1 -type d -name * -execdir mv {} $PROJECTOR_DIR/ide \;
# copy projector files to the container:
COPY static $PROJECTOR_DIR
# copy projector:
# COPY --from=projectorGradleBuilder $PROJECTOR_DIR/projector-server/projector-server/build/distributions/projector-server.zip $PROJECTOR_DIR
# prepare IDE - apply projector-server:
ARG projectorServerUrl
RUN wget -q $projectorServerUrl -O $PROJECTOR_DIR/projector-server.zip && \
    unzip $PROJECTOR_DIR/projector-server.zip && \
    rm $PROJECTOR_DIR/projector-server.zip && \
    find . -maxdepth 1 -type d -name projector-server-* -exec mv {} projector-server \; && \
    mv projector-server $PROJECTOR_DIR/ide/projector-server && \
    mv $PROJECTOR_DIR/ide-projector-launcher.sh $PROJECTOR_DIR/ide/bin && \
    chmod 644 $PROJECTOR_DIR/ide/projector-server/lib/*

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai
# copy the Projector dir:
ENV PROJECTOR_DIR /projector
# COPY --from=projectorStaticFiles $PROJECTOR_DIR $PROJECTOR_DIR

ENV PROJECTOR_USER_NAME smartide

RUN true \
# Any command which returns non-zero exit code will cause this shell script to exit immediately:
    && set -e \
# Activate debugging to show execution details: all commands will be printed before execution
    && set -x \
# move run scipt:
    && mv $PROJECTOR_DIR/run.sh run.sh \
# change user to non-root (http://pjdietz.com/2016/08/28/nginx-in-docker-without-root.html):
    # && mv $PROJECTOR_DIR/$PROJECTOR_USER_NAME /home \
    && chmod g+rw /home && mkdir -p /home/smartide && mkdir -p /home/project \
    # && useradd -d /home/$PROJECTOR_USER_NAME -s /bin/bash -G sudo $PROJECTOR_USER_NAME \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/$PROJECTOR_USER_NAME \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/project \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME $PROJECTOR_DIR/ide/bin \
    && ln -s /projector/ide/bin/idea.sh /usr/local/bin/idea \
    && chmod +x run.sh && chmod +x /projector/ide/bin/ide-projector-launcher.sh

USER $PROJECTOR_USER_NAME
ENV HOME /home/$PROJECTOR_USER_NAME

# use sudo so that user does not get sudo usage info on (the first) login
RUN sudo echo "Running 'sudo' for smartide: success" && \
    # create .bashrc.d folder and source it in the bashrc
    mkdir -p /home/smartide/.bashrc.d && \
    (echo; echo "for i in \$(ls -A \$HOME/.bashrc.d/); do source \$HOME/.bashrc.d/\$i; done"; echo) >> /home/smartide/.bashrc


EXPOSE 8887

#CMD ["bash", "-c", "/run.sh"]

USER root

EXPOSE 8887

ENV HOME /root
RUN echo "export LC_ALL=C.UTF-8">> ~/.bashrc
RUN echo "export LANG=C.UTF-8">> ~/.bashrc
RUN echo "set -g mouse on">> ~/.tmux.conf
##CMD ["bash", "-c", "/run.sh"]

COPY gosu_entrypoint.sh /idesh/gosu_entrypoint.sh
RUN chmod +x /idesh/gosu_entrypoint.sh

ENTRYPOINT ["/idesh/gosu_entrypoint.sh"]
# build.sh
# TAG=2023101602_idea
# REGISTRY=registry.cn-hangzhou.aliyuncs.com/hylstudio/webide-idea
# downloadUrl=https://download.jetbrains.com/idea/ideaIC-2023.1.tar.gz
# projectorServerUrl=https://github.com/956237586/projector-client/releases/download/v1.8.0.2/projector-server-v1.8.1.2.zip
# docker build -f Dockerfile --build-arg downloadUrl=$downloadUrl --build-arg projectorServerUrl=$projectorServerUrl -t $REGISTRY:$TAG .

# start.sh
# docker run -d -e LOCAL_USER_UID=0 \
# -e PERSISTENT_HOME=1 -e PERSISTENT_HOME_DIR=/data/root \
# -e PROJECTOR_SERVER_TOKEN=123456 \
# -p2222:22 \
# -p8887:8887 \
# -v /data:/data \
# registry.cn-hangzhou.aliyuncs.com/hylstudio/webide-idea:2023101602_idea
