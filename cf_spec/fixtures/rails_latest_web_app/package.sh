#!/bin/bash -l

rvm install 2.2.2
rvm use 2.2.2
gem install bundler
bundle package --all
