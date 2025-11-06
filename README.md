# gh-qpr
Quality Pull Requests

## Overview
This is a GitHub CLI extension.  It allows you to create Quality Pull Requests using predefined templates.  You can use the templates from this repo, or create your own!


# Installation

`gh extension install karldreher/gh-qpr`


# Usage
Example: Create a Quality Pull Request

`gh qpr create --template reviewer-first.md`

- Template is always required.  
- In the example above, the `reviewer-first.md` template is selected, from the [`reviewer-first.md`](templates/reviewer-first.md) file in this repository.  (the default repository)

Example: Customize Title
`gh qpr create --template reviewer-first.md --title "feat: new api endpoint`

- Creates the PR as in the previous example, using a Conventional Commit title.

Example: Get Help
`gh qpr` or `gh qpr --help`



# Usage (Advanced)
## Bring Your Own Repo
Set the environment variable `GH_QPR_REPO` in the format `owner/repo`.  This repo will be the place where *all* templates are gathered.  Place them in a directory called `templates` in this repo.

