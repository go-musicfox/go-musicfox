# Easing Functions in Go

### Installation

    go get -u github.com/fogleman/ease

### Documentation

[https://godoc.org/github.com/fogleman/ease](https://godoc.org/github.com/fogleman/ease)

### Usage

All easing functions take a `float64` and return a `float64`. The input should be between 0 and 1, inclusive.

    t = ease.OutElastic(t)
    
Some easing functions have extra parameters, like `period`. Here is an example:

    var f ease.Function
    f = ease.OutElasticFunction(0.5)
    t = f(t)
    
Or, simply...

    t = ease.OutElasticFunction(0.5)(t)

---

![Montage](https://www.michaelfogleman.com/static/ease/montage.png)

---

### ease.Linear(t)
![Linear](https://www.michaelfogleman.com/static/ease/Linear.gif)

---

### ease.InQuad(t)
![InQuad](https://www.michaelfogleman.com/static/ease/InQuad.gif)

### ease.InCubic(t)
![InCubic](https://www.michaelfogleman.com/static/ease/InCubic.gif)

### ease.InQuart(t)
![InQuart](https://www.michaelfogleman.com/static/ease/InQuart.gif)

### ease.InQuint(t)
![InQuint](https://www.michaelfogleman.com/static/ease/InQuint.gif)

### ease.InSine(t)
![InSine](https://www.michaelfogleman.com/static/ease/InSine.gif)

### ease.InExpo(t)
![InExpo](https://www.michaelfogleman.com/static/ease/InExpo.gif)

### ease.InCirc(t)
![InCirc](https://www.michaelfogleman.com/static/ease/InCirc.gif)

### ease.InElastic(t)
![InElastic](https://www.michaelfogleman.com/static/ease/InElastic.gif)

### ease.InBack(t)
![InBack](https://www.michaelfogleman.com/static/ease/InBack.gif)

### ease.InBounce(t)
![InBounce](https://www.michaelfogleman.com/static/ease/InBounce.gif)

---

### ease.OutQuad(t)
![OutQuad](https://www.michaelfogleman.com/static/ease/OutQuad.gif)

### ease.OutCubic(t)
![OutCubic](https://www.michaelfogleman.com/static/ease/OutCubic.gif)

### ease.OutQuart(t)
![OutQuart](https://www.michaelfogleman.com/static/ease/OutQuart.gif)

### ease.OutQuint(t)
![OutQuint](https://www.michaelfogleman.com/static/ease/OutQuint.gif)

### ease.OutSine(t)
![OutSine](https://www.michaelfogleman.com/static/ease/OutSine.gif)

### ease.OutExpo(t)
![OutExpo](https://www.michaelfogleman.com/static/ease/OutExpo.gif)

### ease.OutCirc(t)
![OutCirc](https://www.michaelfogleman.com/static/ease/OutCirc.gif)

### ease.OutElastic(t)
![OutElastic](https://www.michaelfogleman.com/static/ease/OutElastic.gif)

### ease.OutBack(t)
![OutBack](https://www.michaelfogleman.com/static/ease/OutBack.gif)

### ease.OutBounce(t)
![OutBounce](https://www.michaelfogleman.com/static/ease/OutBounce.gif)

---

### ease.InOutQuad(t)
![InOutQuad](https://www.michaelfogleman.com/static/ease/InOutQuad.gif)

### ease.InOutCubic(t)
![InOutCubic](https://www.michaelfogleman.com/static/ease/InOutCubic.gif)

### ease.InOutQuart(t)
![InOutQuart](https://www.michaelfogleman.com/static/ease/InOutQuart.gif)

### ease.InOutQuint(t)
![InOutQuint](https://www.michaelfogleman.com/static/ease/InOutQuint.gif)

### ease.InOutSine(t)
![InOutSine](https://www.michaelfogleman.com/static/ease/InOutSine.gif)

### ease.InOutExpo(t)
![InOutExpo](https://www.michaelfogleman.com/static/ease/InOutExpo.gif)

### ease.InOutCirc(t)
![InOutCirc](https://www.michaelfogleman.com/static/ease/InOutCirc.gif)

### ease.InOutElastic(t)
![InOutElastic](https://www.michaelfogleman.com/static/ease/InOutElastic.gif)

### ease.InOutBack(t)
![InOutBack](https://www.michaelfogleman.com/static/ease/InOutBack.gif)

### ease.InOutBounce(t)
![InOutBounce](https://www.michaelfogleman.com/static/ease/InOutBounce.gif)
