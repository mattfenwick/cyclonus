package utils

import (
	"encoding/json"

	"github.com/mattfenwick/collections/pkg/file"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func JsonStringNoIndent(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	DoOrDie(errors.Wrapf(err, "unable to marshal json"))
	return string(bytes)
}

func ParseYaml[T any](bs []byte) (*T, error) {
	var t T
	if err := yaml.Unmarshal(bs, &t); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal yaml")
	}
	return &t, nil
}

func ParseYamlStrict[T any](bs []byte) (*T, error) {
	var t T
	if err := yaml.UnmarshalStrict(bs, &t); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal yaml")
	}
	return &t, nil
}

func ParseYamlFromFile[T any](path string) (*T, error) {
	bytes, err := file.Read(path)
	if err != nil {
		return nil, err
	}
	return ParseYaml[T](bytes)
}

func ParseYamlFromFileStrict[T any](path string) (*T, error) {
	bytes, err := file.Read(path)
	if err != nil {
		return nil, err
	}
	return ParseYamlStrict[T](bytes)
}

func YamlString(obj interface{}) string {
	bytes, err := yaml.Marshal(obj)
	DoOrDie(errors.Wrapf(err, "unable to marshal yaml"))
	return string(bytes)
}
