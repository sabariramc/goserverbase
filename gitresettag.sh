export tagname="v4.12.0.ddtrace"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname