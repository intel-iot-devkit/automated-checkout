#!/bin/bash
# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

set -e

DIR=$1
MQTT=$2
CONFIDENCE=$3
SKUS=$4


source /opt/intel/openvino_2021/bin/setupvars.sh

/go/src/ds-cv-inference/ds-cv-inference -dir $DIR -mqtt $MQTT -skuMapping $SKUS -model /go/src/ds-cv-inference/product-detection-0001/FP32/product-detection-0001.bin -config /go/src/ds-cv-inference/product-detection-0001/FP32/product-detection-0001.xml -confidence $CONFIDENCE
