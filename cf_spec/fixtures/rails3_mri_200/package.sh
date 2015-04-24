#!/bin/bash -l

rvm install 2.0.0
rvm use 2.0.0
gem install bundler
bundle package --all
