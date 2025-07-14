/*
Package staticlint предоставляет статический анализатор кода (multichecker) для проекта shortygopher.

Multichecker объединяет несколько анализаторов для комплексной проверки кода:

Стандартные анализаторы golang.org/x/tools/go/analysis/passes

# Анализаторы класса SA из staticcheck.io

Анализаторы других классов staticcheck.io:
  - S1xxx (Simple): упрощения кода
  - ST1xxx (StyleCheck): стилистические проблемы
  - QF1xxx (QuickFix): быстрые исправления

Публичные анализаторы:
  - errcheck: проверяет неиспользуемые ошибки
  - shadow: проверяет затенение переменных

Собственный анализатор:
  - osexit: запрещает прямые вызовы os.Exit в функции main пакета main

Использование:

	go run cmd/staticlint/*.go [флаги] [пакеты...]

	или

	go build -o bin/staticlint ./cmd/staticlint
	./bin/staticlint [флаги] [пакеты...]

Примеры запуска:

	# Анализ текущего пакета
	go run cmd/staticlint/*.go .

	# Анализ конкретного пакета
	go run cmd/staticlint/*.go ./internal/app/handlers

	# Анализ всех пакетов проекта
	go run cmd/staticlint/*.go ./...

	# Сборка и запуск через бинарный файл
	go build -o bin/staticlint ./cmd/staticlint
	./bin/staticlint ./...

	# Показать справку
	go run cmd/staticlint/*.go -help

	# Показать справку по конкретному анализатору
	go run cmd/staticlint/*.go -help osexit

Multichecker поддерживает все стандартные флаги анализаторов:

	-c int: количество одновременно выполняемых анализаторов
	-cpuprofile string: записать CPU профиль в файл
	-memprofile string: записать memory профиль в файл
	-json: вывод в формате JSON
	-v: подробный вывод

Exit codes:

	0: анализ завершен без ошибок
	1: обнаружены проблемы в коде
	2: ошибка выполнения анализатора
*/
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/kisielk/errcheck/errcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Add standard analyzers from golang.org/x/tools/go/analysis/passes
	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	// Add all SA analyzers from staticcheck.io
	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Add other classes analyzers from staticcheck.io
	// S1xxx (Simple) - code simplifications
	for _, analyzer := range simple.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// ST1xxx (StyleCheck) - stylistic issues
	for _, analyzer := range stylecheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// QF1xxx (QuickFix) - quick fixes
	for _, analyzer := range quickfix.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// Add public analyzers
	// errcheck - checks for unused errors
	analyzers = append(analyzers, errcheck.Analyzer)

	// shadow - checks for variable shadowing
	analyzers = append(analyzers, shadow.Analyzer)

	// Add our own analyzer
	analyzers = append(analyzers, OSExitAnalyzer)

	// Run multichecker
	multichecker.Main(analyzers...)
}
