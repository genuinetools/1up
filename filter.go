package main

import (
	"github.com/jbrukh/bayesian"
)

const (
	good bayesian.Class = "Good"
	bad  bayesian.Class = "Bad"
)

type classifier struct {
	classifier *bayesian.Classifier
}

func trainClassifier(goodStuff, badStuff []string) (*classifier, error) {
	c := bayesian.NewClassifier(good, bad)

	c.Learn(goodStuff, good)
	c.Learn(badStuff, bad)

	return &classifier{classifier: c}, nil
}

func (c *classifier) classifyMessage(message string) ([]float64, []float64, bool, error) {
	scores, _, _ := c.classifier.LogScores([]string{message})

	// SafeProbScores helps with underflow.
	// See: https://github.com/jbrukh/bayesian/blob/master/bayesian.go#L452
	probs, _, _, err := c.classifier.SafeProbScores([]string{message})
	if err != nil {
		return nil, nil, false, err
	}

	isBad := false
	// If the probability that it is good is less than 40% and
	// the probability that it is bad is greater than 75%, then it
	// is probably bad.
	if probs[0] < 0.4 && probs[1] > 0.75 {
		isBad = true
	}

	return scores, probs, isBad, err
}
