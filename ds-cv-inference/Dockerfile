# Copyright © 2021 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM hybridgroup/gocv:latest as builder

WORKDIR /tmp
RUN wget https://apt.repos.intel.com/openvino/2021/GPG-PUB-KEY-INTEL-OPENVINO-2021
RUN apt-key add GPG-PUB-KEY-INTEL-OPENVINO-2021 && apt-key list
RUN echo "deb https://apt.repos.intel.com/openvino/2021 all main" | tee /etc/apt/sources.list.d/intel-openvino-2021.list
RUN apt update
RUN apt install intel-openvino-dev-ubuntu18-2021.3.394 -y

RUN apt install libgtk-3-dev -y
WORKDIR /go/src/ds-cv-inference

COPY go.* /go/src/ds-cv-inference/
RUN go mod download

COPY . /go/src/ds-cv-inference

RUN /bin/bash -c "source /opt/intel/openvino_2021/bin/setupvars.sh && go build -o ds-cv-inference"
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
