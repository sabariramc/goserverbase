export tagname="v4.12.0"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname