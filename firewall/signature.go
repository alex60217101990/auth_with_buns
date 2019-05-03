package firewall

import (
	"auth_service_template/logger"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type FirewallObj struct {
	visitors sync.Map
	buns     sync.Map
	log      *logger.Log
	env      *map[string]string
}

func NewFirewall(logs *logger.Log, glob_env *map[string]string) *FirewallObj {
	return &FirewallObj{
		log: logs,
		env: glob_env,
	}
}

func (f *FirewallObj) GetVisitor(ip string) *rate.Limiter {
	v, exists := f.visitors.Load(ip)
	if exists == false {
		limiter := rate.NewLimiter(2, 5)
		f.visitors.Store(ip, &visitor{limiter, time.Now()})
		return limiter
	} else {
		return v.(*visitor).limiter
	}
}

func (f *FirewallObj) GetCurrentIp(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return IPAddress
}

func (f *FirewallObj) getCurrentBunInterval() time.Duration {
	var t time.Duration
	if h, ok := (*f.env)["BUN_TIMEOUT_HOUR"]; ok {
		if i32, err := strconv.Atoi(h); err == nil {
			t = t + time.Duration(i32)*time.Hour
		}
	}
	if h, ok := (*f.env)["BUN_TIMEOUT_MIN"]; ok {
		if i32, err := strconv.Atoi(h); err == nil {
			t = t + time.Duration(i32)*time.Minute
		}
	}
	if h, ok := (*f.env)["BUN_TIMEOUT_SEC"]; ok {
		if i32, err := strconv.Atoi(h); err == nil {
			t = t + time.Duration(i32)*time.Second
		}
	}
	return t
}

func (f *FirewallObj) CleanupVisitors(ctx context.Context) {
	var intervalCloseBuns, intervalCleanupVisitors time.Duration
	if i32, err := strconv.Atoi((*f.env)["BUN_INSPECTION_PERIOD"]); err == nil {
		intervalCloseBuns = time.Duration(i32) * time.Second
	} else {
		intervalCloseBuns = time.Duration(30) * time.Second
	}
	//intervalCloseBuns := f.getCurrentInterval()
	tickerCloseBuns := time.NewTicker(intervalCloseBuns)

	if i32, err := strconv.Atoi((*f.env)["CLEANUP_VISITORS_PERIOD"]); err == nil {
		intervalCleanupVisitors = time.Duration(i32) * time.Second
	} else {
		intervalCleanupVisitors = time.Duration(30) * time.Minute
	}
	tickerCleanupVisitors := time.NewTicker(intervalCleanupVisitors)

	defer func() {
		tickerCleanupVisitors.Stop()
		tickerCloseBuns.Stop()
	}()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-tickerCleanupVisitors.C:
			f.visitors.Range(func(ip, v interface{}) bool {
				if time.Now().Sub(v.(*visitor).lastSeen) > 3*time.Minute {
					f.visitors.Delete(ip)
				}
				return true
			})
		case t, ok := <-tickerCloseBuns.C:
			if ok {
				f.buns.Range(func(ip, b interface{}) bool {
					if t.Sub(b.(*bun).startBun) >= f.getCurrentBunInterval() {
						f.buns.Delete(ip)
					}
					return true
				})
			}
		}
	}
}

func (f *FirewallObj) BunHttpMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := f.GetCurrentIp(r)
		if len(ip) > 0 {
			if f.isBun(&ip) {
				http.Error(w, http.StatusText(502), http.StatusBadGateway)
				return
			} else {
				ctx := newContextWithIP(r.Context(), ip)
				next(w, r.WithContext(ctx))
			}
		} else {
			http.Error(w, http.StatusText(502), http.StatusBadGateway)
			return
		}
	})
}

func (f *FirewallObj) isBun(ip *string) bool {
	if obj, ok := f.buns.Load(*ip); ok {
		return obj.(*bun).isBun
	}
	return false
}

func (f *FirewallObj) LimitHttpMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := requestIPFromContext(r.Context())
		limiter := f.GetVisitor(ip)
		fmt.Printf("IP: [%s],\t,\tlimiter-limit: [%v], => [%v]\n",
			ip, limiter.Limit(), limiter.Allow())
		if limiter.Allow() == false {
			if ok := f.incrementFailedRequests(&ip); ok {
				http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
				return
			}
		}
		go f.PrintTables()
		next(w, r)
	})
}

func (f *FirewallObj) incrementFailedRequests(ip *string) bool {
	setBun := false
	if max_attempts, err := strconv.Atoi((*f.env)["MAX_ERR_LIMIT_REQ_ATTEMPTS"]); err == nil {
		if old, ok := f.buns.Load(*ip); ok {
			newBun := bun{}
			newBun.counter = old.(bun).counter + 1
			if newBun.counter >= max_attempts {
				newBun.isBun = true
				newBun.startBun = time.Now()
				f.buns.Store(*ip, &newBun)
				setBun = true
			}
			f.buns.Store(ip, old)
		}
	}
	return setBun
}

func (f *FirewallObj) PrintTables() {
	f.visitors.Range(func(ip, v interface{}) bool {
		fmt.Printf("ip: [%s],\tlastSeen: [%s],\tlimiter-limit: [%d]\n",
			ip.(string), (*v.(*visitor)).lastSeen.Format("2006/01/02 15:04:05"),
			(*v.(*visitor)).limiter.Limit())
		return true
	})
	f.buns.Range(func(ip, b interface{}) bool {
		fmt.Println(map[string]bun{ip.(string): *b.(*bun)})
		return true
	})
}
