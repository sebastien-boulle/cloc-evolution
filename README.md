`cloc-evolution` is a tool for running [http://cloc.sourceforge.net/](cloc) across different tagged release versions in
a git repository. Given a target path pointing to a git repository, it uses [https://github.com/libgit2/git2go](git2go)
to read the tags in the repository. It assumes all tags contain a [http://semver.org/](semver) versioning string (i.e
<number>.<number>.<number>) somewhere inside them. It will then sort these by version number and use git2go to check out
each tag in turn, whereby it runs `cloc` on them. Once it has iterated over all tags it will produce a temporary file
containing an HTML output with a [http://highcharts.com/](highcharts) chart showing the evolution of the number of lines
of code for different languages across version changes.

**Note**: This will destroy any local changes in the git repository. Use at your own risk.

This project is my first time using golang. There may be issues or improvements that can be made. Pull requests or
suggestions are welcome.
