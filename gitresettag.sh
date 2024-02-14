export tagname="v5.0.3"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname