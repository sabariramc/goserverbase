export tagname="v4.6.1"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname