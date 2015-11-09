This directory contains vendored dependencies.

To add a new dependency:

./add-dep.sh <import-path>

To update a dependency:

./update-dep.sh <import-path>

Note that for the moment these scripts only work on git repositories. Also the
import path rewriting is really dumb.
