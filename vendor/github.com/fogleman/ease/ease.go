package ease

import "math"

type Function func(float64) float64

func Linear(t float64) float64 {
	return t
}

func InQuad(t float64) float64 {
	return t * t
}

func OutQuad(t float64) float64 {
	return -t * (t - 2)
}

func InOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	} else {
		t = 2*t - 1
		return -0.5 * (t*(t-2) - 1)
	}
}

func InCubic(t float64) float64 {
	return t * t * t
}

func OutCubic(t float64) float64 {
	t -= 1
	return t*t*t + 1
}

func InOutCubic(t float64) float64 {
	t *= 2
	if t < 1 {
		return 0.5 * t * t * t
	} else {
		t -= 2
		return 0.5 * (t*t*t + 2)
	}
}

func InQuart(t float64) float64 {
	return t * t * t * t
}

func OutQuart(t float64) float64 {
	t -= 1
	return -(t*t*t*t - 1)
}

func InOutQuart(t float64) float64 {
	t *= 2
	if t < 1 {
		return 0.5 * t * t * t * t
	} else {
		t -= 2
		return -0.5 * (t*t*t*t - 2)
	}
}

func InQuint(t float64) float64 {
	return t * t * t * t * t
}

func OutQuint(t float64) float64 {
	t -= 1
	return t*t*t*t*t + 1
}

func InOutQuint(t float64) float64 {
	t *= 2
	if t < 1 {
		return 0.5 * t * t * t * t * t
	} else {
		t -= 2
		return 0.5 * (t*t*t*t*t + 2)
	}
}

func InSine(t float64) float64 {
	return -1*math.Cos(t*math.Pi/2) + 1
}

func OutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

func InOutSine(t float64) float64 {
	return -0.5 * (math.Cos(math.Pi*t) - 1)
}

func InExpo(t float64) float64 {
	if t == 0 {
		return 0
	} else {
		return math.Pow(2, 10*(t-1))
	}
}

func OutExpo(t float64) float64 {
	if t == 1 {
		return 1
	} else {
		return 1 - math.Pow(2, -10*t)
	}
}

func InOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	} else if t == 1 {
		return 1
	} else {
		if t < 0.5 {
			return 0.5 * math.Pow(2, (20*t)-10)
		} else {
			return 1 - 0.5*math.Pow(2, (-20*t)+10)
		}
	}
}

func InCirc(t float64) float64 {
	return -1 * (math.Sqrt(1-t*t) - 1)
}

func OutCirc(t float64) float64 {
	t -= 1
	return math.Sqrt(1 - (t * t))
}

func InOutCirc(t float64) float64 {
	t *= 2
	if t < 1 {
		return -0.5 * (math.Sqrt(1-t*t) - 1)
	} else {
		t = t - 2
		return 0.5 * (math.Sqrt(1-t*t) + 1)
	}
}

func InElastic(t float64) float64 {
	return InElasticFunction(0.5)(t)
}

func OutElastic(t float64) float64 {
	return OutElasticFunction(0.5)(t)
}

func InOutElastic(t float64) float64 {
	return InOutElasticFunction(0.5)(t)
}

func InElasticFunction(period float64) Function {
	p := period
	return func(t float64) float64 {
		t -= 1
		return -1 * (math.Pow(2, 10*t) * math.Sin((t-p/4)*(2*math.Pi)/p))
	}
}

func OutElasticFunction(period float64) Function {
	p := period
	return func(t float64) float64 {
		return math.Pow(2, -10*t)*math.Sin((t-p/4)*(2*math.Pi/p)) + 1
	}
}

func InOutElasticFunction(period float64) Function {
	p := period
	return func(t float64) float64 {
		t *= 2
		if t < 1 {
			t -= 1
			return -0.5 * (math.Pow(2, 10*t) * math.Sin((t-p/4)*2*math.Pi/p))
		} else {
			t -= 1
			return math.Pow(2, -10*t)*math.Sin((t-p/4)*2*math.Pi/p)*0.5 + 1
		}
	}
}

func InBack(t float64) float64 {
	s := 1.70158
	return t * t * ((s+1)*t - s)
}

func OutBack(t float64) float64 {
	s := 1.70158
	t -= 1
	return t*t*((s+1)*t+s) + 1
}

func InOutBack(t float64) float64 {
	s := 1.70158
	t *= 2
	if t < 1 {
		s *= 1.525
		return 0.5 * (t * t * ((s+1)*t - s))
	} else {
		t -= 2
		s *= 1.525
		return 0.5 * (t*t*((s+1)*t+s) + 2)
	}
}

func InBounce(t float64) float64 {
	return 1 - OutBounce(1-t)
}

func OutBounce(t float64) float64 {
	if t < 4/11.0 {
		return (121 * t * t) / 16.0
	} else if t < 8/11.0 {
		return (363 / 40.0 * t * t) - (99 / 10.0 * t) + 17/5.0
	} else if t < 9/10.0 {
		return (4356 / 361.0 * t * t) - (35442 / 1805.0 * t) + 16061/1805.0
	} else {
		return (54 / 5.0 * t * t) - (513 / 25.0 * t) + 268/25.0
	}
}

func InOutBounce(t float64) float64 {
	if t < 0.5 {
		return InBounce(2*t) * 0.5
	} else {
		return OutBounce(2*t-1)*0.5 + 0.5
	}
}

func InSquare(t float64) float64 {
	if t < 1 {
		return 0
	} else {
		return 1
	}
}

func OutSquare(t float64) float64 {
	if t > 0 {
		return 1
	} else {
		return 0
	}
}

func InOutSquare(t float64) float64 {
	if t < 0.5 {
		return 0
	} else {
		return 1
	}
}
