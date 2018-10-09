# Contributing to secrets-manager

If you find something missing or not working as expected, we are happy to receive your pull requests! These are the guidelines to follow to make your awesome code part of _secrets-manager_

## Steps to Contribute

* Create an Issue
* Fork _secrets-manager_ and work on the new feature/bugfix
* Open a Pull Request to propose your changes:
    * New features: into master
    * Bugfixes: into latest release branch

## Git Branches Model

We have 4 kind of branches in _secrets-manager_ development:

* **`master`**: is the integration branch, new features are merged into this. Developers create their feature branch and open a Pull Request to master to propose the changes.
* **`release-*` branches**: are used to prepare every new _secrets-manager_ minor release (`<major>.<minor>`, ie: 0.1, 1.2, etc). In these branches we don't merge new features, only bugfixes. Once a new bug is fixed in the latest release branch, it has to be merged into `master` too.
* **feature branches**: branches created by developers to implement new functinalities in _secrets-manager_. They can only be merged into master.
* **bugfix branches**: branches with fixes for bugs, they can only be merged into release branches.

Given this, we use `master` as our integration branch and maintain separated branches for each minor release. New features should be merged into `master` branch, and bugfixes should be merged into `release-*` branches. For a patch release, work on the corresponding minor release branch.

### Publishing the release

Once the work in the release branch is stabilized, create a tag in the branch.

### Release Candidates

Release candidates are treated as normal releases, but they must append `-rc[0-9]*` to the branch name.