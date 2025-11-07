# gh-qpr
Quality Pull Requests

## Overview
This is a GitHub CLI extension.  It allows you to create Quality Pull Requests using predefined templates.  You can use the templates from this repo, or create your own!

## Background
Pull requests (PRs) in software development become the most important medium for communication.  Especially as AI accelerates the *generation of code*, the *sharing of context* between engineers remains *highly important*.  A high volume of code might not even be able to be reviewed on a line-by-line basis in the near future, but at minimum, *humans should be able to clearly communicate change scope to other humans* for PRs to remain relevant.  

### References
Other articles affirm this point, and provide great guidance on what makes up a good pull request. 

https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/getting-started/helping-others-review-your-changes

https://www.atlassian.com/blog/git/written-unwritten-guide-pull-requests

https://github.com/scastiel/book-pr/blob/main/manuscript.md

https://www.5xx.engineer/2025/02/13/rfprs.html

### Quality Pull Requests

This tool provides an interface, which: 
- Promotes the practice of creating good pull requests
- Makes creating full-featured, clearly written pull requests a repeatable practice
- Allows users to maintain templates that are useful to them (as individuals, transcending their contributorship to [specific repos and orgs](https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/creating-a-pull-request-template-for-your-repository))


# Installation

`gh extension install karldreher/gh-qpr`


# Usage
Example: Create a Quality Pull Request

`gh qpr create --template reviewer-first.md`

- Template is always required.  
- In the example above, the `reviewer-first.md` template is selected, from the [`reviewer-first.md`](templates/reviewer-first.md) file in this repository.  (the default repository)
- `.md` extensions are optional in template selection.

Example: Customize Title
`gh qpr create --template reviewer-first --title "feat: new api endpoint`

- Creates the PR as in the previous example, using a Conventional Commit title.

Example: Get Help
`gh qpr` or `gh qpr --help`



# Usage (Advanced)
## Bring Your Own Repo
Set the environment variable `GH_QPR_REPO` in the format `owner/repo`.  This repo will be the place where *all* templates are gathered.  Place them in a directory called `templates` in this repo.

