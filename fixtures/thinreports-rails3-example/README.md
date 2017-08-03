# Thinreports Rails3 Example

The Simple Task Management Application using Thinreports and Rails3. 
**Rails4 example** is [here](https://github.com/thinreports/thinreports-rails4-example).

[![Build Status](http://img.shields.io/travis/thinreports/thinreports-rails3-example.svg?style=flat)](https://travis-ci.org/thinreports/thinreports-rails3-example)
[![Dependency Status](http://img.shields.io/gemnasium/thinreports/thinreports-rails3-example.svg?style=flat)](https://gemnasium.com/thinreports/thinreports-rails3-example)

## How to run this example:

Get this application source using git:

    $ git clone git://github.com/thinreports/thinreports-rails3-example.git

Or download ZIP archives from [here](https://github.com/thinreports/thinreports-rails3-example/archive/master.zip).

Then move to application directory, and bundle:

    $ cd thinreports-rails3-example/
    $ bundle install

Setup database with seeds:

    $ bundle exec rake db:setup

Start application:

    $ bundle exec rails s

Go to `http://localhost:3000/tasks` in your browser.

### Requirements

* Ruby 1.9.3, 2.0, 2.1.2
* Rails 3.2.19
* thinreports 0.7.7
* thinreports-rails 0.1.3

## Development

### How to run the test

    $ bundle exec rake spec

## Copyright

&copy; 2010-2015 [Matsukei Co.,Ltd](http://www.matsukei.co.jp).
