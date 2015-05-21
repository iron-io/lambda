# How to make a build for Windows

### 1\. Build iron.exe

If you never performed it, [prepare your Go-Envireonment](build_go_for_windows.md).

After it just build an exe:

```sh
bin/windows$ GOOS=windows GOARCH=amd64 go build -o iron.exe ../../*.go
```

### 2\. Update IronCLI_README.pdf

```sh
bin/windows$ curl 'https://gitprint.com/iron-io/ironcli/blob/master/README.md?download' > IronCLI_README.pdf
```

### 3\. Build MSI file

[Install WIX](http://wixtoolset.org/releases/) on your Windows machine (or anything else + Wine)

Update version of the installation. 

* Replace 2 GUIDs in `<Product />` section with a couple of new
* Increase `Version` attribute inside `<Product />`

After it perform 2 easy steps on Windows machine (or anything else + Wine):

```
candle iron.wxs
light iron.wixobj
```

After it `iron.msi` should appear in the folder.

### 4\. Upload it to Github Releases
