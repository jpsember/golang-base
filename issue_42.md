# Configuring repo to not merge particular files

Adapted from: https://medium.com/@porteneuve/how-to-make-git-preserve-specific-files-while-merging-18c92343826b

I ran this command:
```
git config merge.ours.driver true
```

I created a file `.gitattributes` (and committed it) with this content:
```
webapp/database.go merge=ours
```

