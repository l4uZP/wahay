//go:build !windows

package tor

import log "github.com/sirupsen/logrus"

func isThereConfiguredTorBinary(path string) (b *binary, err error) {
	if len(path) == 0 {
		return b, ErrInvalidTorPath
	}

	if !filesystemf.IsADirectory(path) {
		b, err = getBinaryForPath(path)
		return
	}

	list := listPossibleTorBinary(path)

	if len(list) > 0 {
		for _, p := range list {
			b, _ = getBinaryForPath(p)
			if b.isValid {
				return b, nil
			}
		}
	}

	return
}

func findTorBinaryInSystem() (b *binary, fatalErr error) {
	path, err := execf.LookPath("tor")
	if err != nil {
		return nil, nil
	}

	log.Debugf("findTorBinaryInSystem(%s)", path)

	b, errTorBinary := isThereConfiguredTorBinary(path)
	if errTorBinary != nil {
		return nil, nil
	}

	return b, nil
}
