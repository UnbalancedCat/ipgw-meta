package casbundle

func OpenQRDevelopment(directory, workerHash, inputsHash, productHash string) (*AuthDevelopment, error) {
	names := []string{"build-inputs.json", "ipgw-meta.exe", "qr-dev.exe"}
	hashes := []string{inputsHash, productHash, workerHash}
	for _, h := range hashes {
		if !digestPattern.MatchString(h) {
			return nil, ErrInvalid
		}
	}
	b, paths, err := openAuthFiles(directory, names)
	if err != nil {
		return nil, ErrInvalid
	}
	d := &AuthDevelopment{b: b, names: names, hashes: hashes, ProductPath: paths[1], ObserverPath: paths[2]}
	if d.Verify() != nil {
		d.Close()
		return nil, ErrInvalid
	}
	return d, nil
}
