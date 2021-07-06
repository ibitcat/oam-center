#!/bin/bash

#go build -ldflags "-w -s"
#go build -v -x

zip -vr oam-center.zip oam-center conf.yaml.tpl views/ static/ *.sql -x "conf.yaml"
