#!/bin/bash -l

rvm install 2.2.0
rvm use 2.2.0
bundle package --all
