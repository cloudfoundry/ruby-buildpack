#!/usr/bin/env bash -l

rvm use <%=engine%>-<%= engine_version %> --install > /dev/null 2>&1
gem install bundler -v '1.8.0'
bundle '_1.8.0_' package --all
