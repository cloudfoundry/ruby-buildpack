package versions_test

import "strings"

const windowsOnlyGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  x64-mingw32
  mswin32

DEPENDENCIES
  sinatra
`

var windowsEndingsGemfileLockFixture = strings.Join([]string{"GEM",
	"  remote: https://rubygems.org/",
	"  specs:",
	"    rack (1.5.2)",
	"    rack-protection (1.5.2)",
	"      rack",
	"    sinatra (1.4.4)",
	"      rack (~> 1.4)",
	"      rack-protection (~> 1.4)",
	"      tilt (~> 1.3, >= 1.3.4)",
	"    tilt (1.4.1)",
	"",
	"PLATFORMS",
	"  x64-mingw32",
	"  ruby",
	"",
	"DEPENDENCIES",
	"  sinatra"},
	"\r\n")

const rubyGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  ruby

DEPENDENCIES
  sinatra
`

const jrubyGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  jruby

DEPENDENCIES
  sinatra
`

const bothGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  x64-mingw32
  ruby

DEPENDENCIES
  sinatra
`

const linuxGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  x86_64-linux

DEPENDENCIES
  sinatra
`

const mingwLinuxGemfileLockFixture = `GEM
  remote: https://rubygems.org/
  specs:
    rack (1.5.2)
    rack-protection (1.5.2)
      rack
    sinatra (1.4.4)
      rack (~> 1.4)
      rack-protection (~> 1.4)
      tilt (~> 1.3, >= 1.3.4)
    tilt (1.4.1)

PLATFORMS
  x64-mingw32
  x86_64-linux

DEPENDENCIES
  sinatra
`
