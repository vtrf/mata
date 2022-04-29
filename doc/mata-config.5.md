---
title: mata-config
section: 5
header: User Manual
date: April 26, 2022
---

# Name

mata-config - configuration file formats for *mata*(1)

# Configuration

There is only one configuration file for mata: *config.json*. The program looks 
after this file in your XDG config home plus "mata", which defaults to 
~/.config/mata or the path on *MATAROA_CONFIG*. This file uses the _json_ format.

Another way to configure the CLI is to set the respective *MATAROA_ENDPOINT* and
*MATAROA_KEY* environment variables. If both variables are set, the program will
skip reading the *config.json* file. This is useful to use this tool in a 
Continuous Integration environment.

# CONFIG.JSON

This file is used to configure the behavior of mata.

## OPTIONS

**endpoint** 
    Sets the endpoint that will be used to send resquests.
    Default: https://mataroa.blog/api

**key**
    The API key provided by your mataroa dashboard.

# SEE ALSO

*mata*(1)

# AUTHORS

Created by Victor Freire <victor@freire.dev.br>. For more information about mata
development, see https://sr.ht/~glorifiedgluer/mata.