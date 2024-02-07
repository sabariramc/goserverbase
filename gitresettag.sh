export tagname="v4.14.2"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname