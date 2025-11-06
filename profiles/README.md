Описание оптимизиации:
zap.SugaredLogger был заменен на структурированный zap.Logger в app.go. Это позволило снизить потребление памяти на 11%. 
zap.SugaredLogger использует fmt.Sprintf-подобный интерфейс, который создает временные строки. Структурированный
логгер работает без промежуточного форматирования и с меньшим количеством аллокаций, что и позволило сократить потребление памяти.


Type: inuse_space
Time: 2025-11-01 19:27:25 MSK
Showing nodes accounting for -1027.49kB, 11.13% of 9229.88kB total
Dropped 8 nodes (cum <= 46.15kB)
      flat  flat%   sum%        cum   cum%
 1025.12kB 11.11% 11.11%  1025.12kB 11.11%  runtime.makeProfStackFP (inline)
 1024.11kB 11.10% 22.20%  1024.11kB 11.10%  context.(*cancelCtx).Done
-1024.02kB 11.09% 11.11% -1024.02kB 11.09%  internal/syscall/windows.errnoErr (inline)
    -514kB  5.57%  5.54%     -514kB  5.57%  bufio.NewReaderSize (inline)
    -513kB  5.56% 0.019%     -513kB  5.56%  internal/syscall/windows/registry.Key.ReadSubKeyNames
    -513kB  5.56%  5.58%  1024.25kB 11.10%  runtime.allocm
 -512.50kB  5.55% 11.13%  -512.50kB  5.55%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
 -512.31kB  5.55% 16.68%  -512.31kB  5.55%  net.newFD (inline)
  512.12kB  5.55% 11.13%  1537.25kB 16.66%  runtime.mcommoninit
 -512.12kB  5.55% 16.68%  -512.12kB  5.55%  github.com/go-chi/chi/v5.NewRouteContext
  512.12kB  5.55% 11.13%  2049.23kB 22.20%  net/http.(*conn).readRequest
 -512.03kB  5.55% 16.68%  -512.03kB  5.55%  fmt.Sprintf
  512.02kB  5.55% 11.13%   512.02kB  5.55%  github.com/go-chi/chi/v5.(*node).FindRoute
         0     0% 11.13%     -514kB  5.57%  bufio.NewReader (inline)
         0     0% 11.13%  1024.11kB 11.10%  context.(*cancelCtx).propagateCancel
         0     0% 11.13%  1024.11kB 11.10%  context.WithCancel
         0     0% 11.13%  1024.11kB 11.10%  context.withCancel (inline)
         0     0% 11.13%  -512.60kB  5.55%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 11.13% -1025.31kB 11.11%  github.com/noedaka/go-url-shortener/internal/app.Run
         0     0% 11.13%  -512.12kB  5.55%  github.com/noedaka/go-url-shortener/internal/app.Run.NewRouter.NewMux.func3
         0     0% 11.13%   512.02kB  5.55%  github.com/noedaka/go-url-shortener/internal/app.Run.func1.AuditMiddleware.2.1
         0     0% 11.13%   512.02kB  5.55%  github.com/noedaka/go-url-shortener/internal/middleware.AuthMiddleware.func1
         0     0% 11.13%   512.02kB  5.55%  github.com/noedaka/go-url-shortener/internal/middleware.GzipMiddleware.func1
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap.(*Logger).Info
         0     0% 11.13%     -513kB  5.56%  go.uber.org/zap.(*SugaredLogger).Infof
         0     0% 11.13%     -513kB  5.56%  go.uber.org/zap.(*SugaredLogger).log
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap/buffer.Pool.Get
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap/internal/bufferpool.init.NewPool.New[go.shape.*uint8].func2
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 11.13% -1025.50kB 11.11%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 11.13% -1025.50kB 11.11%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap/zapcore.EntryCaller.TrimmedPath
         0     0% 11.13%     -513kB  5.56%  go.uber.org/zap/zapcore.ISO8601TimeEncoder
         0     0% 11.13%  -512.50kB  5.55%  go.uber.org/zap/zapcore.ShortCallerEncoder
         0     0% 11.13% -1025.50kB 11.11%  go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry
         0     0% 11.13%     -513kB  5.56%  go.uber.org/zap/zapcore.encodeTimeLayout
         0     0% 11.13% -1024.02kB 11.09%  internal/poll.(*FD).Read
         0     0% 11.13% -1024.02kB 11.09%  internal/poll.execIO
         0     0% 11.13% -1024.02kB 11.09%  internal/syscall/windows.WSAGetOverlappedResult
         0     0% 11.13% -1025.31kB 11.11%  main.main
         0     0% 11.13%  -512.31kB  5.55%  net.(*TCPListener).Accept
         0     0% 11.13%  -512.31kB  5.55%  net.(*TCPListener).accept
         0     0% 11.13% -1024.02kB 11.09%  net.(*conn).Read
         0     0% 11.13% -1024.02kB 11.09%  net.(*netFD).Read
         0     0% 11.13%  -512.31kB  5.55%  net.(*netFD).accept
         0     0% 11.13%  -512.03kB  5.55%  net/http.(*ServeMux).register (inline)
         0     0% 11.13%  -512.03kB  5.55%  net/http.(*ServeMux).registerErr
         0     0% 11.13%  -512.31kB  5.55%  net/http.(*Server).ListenAndServe
         0     0% 11.13%  -512.31kB  5.55%  net/http.(*Server).Serve
         0     0% 11.13%   508.63kB  5.51%  net/http.(*conn).serve
         0     0% 11.13% -1024.02kB 11.09%  net/http.(*connReader).backgroundRead
         0     0% 11.13%  -512.03kB  5.55%  net/http.HandleFunc
         0     0% 11.13%  -512.31kB  5.55%  net/http.ListenAndServe (inline)
         0     0% 11.13%     -514kB  5.57%  net/http.newBufioReader
         0     0% 11.13%  -512.60kB  5.55%  net/http.serverHandler.ServeHTTP
         0     0% 11.13%  -512.03kB  5.55%  net/http/pprof.init.0
         0     0% 11.13%  -512.03kB  5.55%  runtime.doInit (inline)
         0     0% 11.13%  -512.03kB  5.55%  runtime.doInit1
         0     0% 11.13%   512.56kB  5.55%  runtime.handoffp
         0     0% 11.13%  1025.12kB 11.11%  runtime.mProfStackInit (inline)
         0     0% 11.13% -1537.34kB 16.66%  runtime.main
         0     0% 11.13%     -513kB  5.56%  runtime.mcall
         0     0% 11.13%  2049.81kB 22.21%  runtime.mstart
         0     0% 11.13%  2049.81kB 22.21%  runtime.mstart0
         0     0% 11.13%  2049.81kB 22.21%  runtime.mstart1
         0     0% 11.13%  1024.25kB 11.10%  runtime.newm
         0     0% 11.13%     -513kB  5.56%  runtime.park_m
         0     0% 11.13%  -512.56kB  5.55%  runtime.ready
         0     0% 11.13%  -512.56kB  5.55%  runtime.readyWithTime.goready.func1
         0     0% 11.13%  1024.25kB 11.10%  runtime.resetspinning
         0     0% 11.13%   512.56kB  5.55%  runtime.retake
         0     0% 11.13%  1024.25kB 11.10%  runtime.schedule
         0     0% 11.13%  1024.25kB 11.10%  runtime.startm
         0     0% 11.13%   512.56kB  5.55%  runtime.sysmon
         0     0% 11.13%  -512.56kB  5.55%  runtime.systemstack
         0     0% 11.13%   511.69kB  5.54%  runtime.wakep
         0     0% 11.13%     -513kB  5.56%  sync.(*Once).Do (inline)
         0     0% 11.13%     -513kB  5.56%  sync.(*Once).doSlow
         0     0% 11.13% -1024.62kB 11.10%  sync.(*Pool).Get
         0     0% 11.13%     -513kB  5.56%  time.(*Location).get
         0     0% 11.13%     -513kB  5.56%  time.Time.AppendFormat
         0     0% 11.13%     -513kB  5.56%  time.Time.Format
         0     0% 11.13%     -513kB  5.56%  time.Time.appendFormat
         0     0% 11.13%     -513kB  5.56%  time.Time.locabs
         0     0% 11.13%     -513kB  5.56%  time.abbrev
         0     0% 11.13%     -513kB  5.56%  time.initLocal
         0     0% 11.13%     -513kB  5.56%  time.initLocalFromTZI
         0     0% 11.13%     -513kB  5.56%  time.toEnglishName