package types

import "html/template"

var TPLIndex = template.Must(
	template.New("index").Parse(
		`
<!DOCTYPE html>
<html lang="ru"><head>
  <meta charset="UTF-8"><title>Одностаночное планирование</title>
  <style>
    body{font-family:Arial,sans-serif;background:#f5f6fa;margin:0;padding:0}
    .container{max-width:600px;margin:50px auto;background:#fff;border-radius:8px;
      box-shadow:0 2px 8px rgba(0,0,0,0.1);padding:20px}
    h1{text-align:center;color:#0984e3;margin-bottom:20px}
    form{display:grid;gap:15px}
    label{display:flex;flex-direction:column;font-weight:bold;color:#2d3436}
    input[type=number]{margin-top:5px;padding:8px;border:1px solid #dcdde1;border-radius:4px}
    button{padding:12px;background:#0984e3;color:#fff;border:none;border-radius:4px;
      font-size:16px;cursor:pointer;transition:background .2s}
    button:hover{background:#74b9ff}
    .generate{ text-align:center;margin:30px 0 10px;color:#636e72;font-size:14px}
  </style>
</head><body>
  <div class="container">
    <h1>Одностаночное планирование</h1>

    <!-- Ручной ввод -->
    <form method="POST" action="/params">
      <label>Число типов (T):
        <input type="number" name="numTypes" min="1" required>
      </label>
      <label>Копий каждого типа (P):
        <input type="number" name="piecesPerType" min="1" required>
      </label>
      <button type="submit">Ввести вручную →</button>
    </form>

    <div class="generate">— или —</div>

    <!-- Генерация и сразу решение -->
    <form method="POST" action="/compute">
      <input type="hidden" name="generate" value="1">
      <label>Число типов (T):
        <input type="number" name="numTypes" min="1" required>
      </label>
      <label>Копий каждого типа (P):
        <input type="number" name="piecesPerType" min="1" required>
      </label>

      <h3>Параметры генератора</h3>
      <label>Min Setup:
        <input type="number" name="minSetup" min="0" value="1" required>
      </label>
      <label>Max Setup:
        <input type="number" name="maxSetup" min="0" value="10" required>
      </label>
      <label>Min Process:
        <input type="number" name="minProcess" min="0" value="1" required>
      </label>
      <label>Max Process:
        <input type="number" name="maxProcess" min="0" value="15" required>
      </label>

      <button type="submit">Сгенерировать и решить</button>
    </form>
  </div>
</body>
</html>
`,
	),
)

var TPLParams = template.Must(
	template.New("params").Funcs(
		template.FuncMap{
			"add": func(a, b int) int { return a + b },
		},
	).Parse(
		`<!DOCTYPE html>
<html lang="ru"><head><meta charset="UTF-8"><title>Параметры задач</title>
<style>body{font-family:Arial,sans-serif;background:#f5f6fa;margin:0;padding:0}
.container{max-width:800px;margin:30px auto;background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1);padding:20px}
h1{text-align:center;color:#0984e3}h2{margin-top:30px;color:#2d3436;border-bottom:2px solid #dcdde1;padding-bottom:5px}
fieldset{border:1px solid #dcdde1;border-radius:4px;padding:15px;margin-bottom:20px;background:#f0f0f0}
legend{font-weight:bold}label{display:block;margin:10px 0 5px}
input[type=number]{width:100%;padding:6px;border:1px solid #dcdde1;border-radius:4px}
.col-group{display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:10px}
button{width:100%;padding:10px;background:#0984e3;color:#fff;border:none;border-radius:4px;font-size:16px;cursor:pointer}
button:hover{background:#74b9ff}</style>
</head><body><div class="container">
<h1>Параметры для {{.T}} типов и {{.P}} копий</h1>
<form method="POST" action="/compute">
  <input type="hidden" name="numTypes" value="{{.T}}">
  <input type="hidden" name="piecesPerType" value="{{.P}}">
  <h2>Время переналадки (Setup)</h2><div class="col-group">
    {{range $i := .RangeT}}<label>Type {{add $i 1}}:<input type="number" name="setup{{$i}}" min="0" required></label>{{end}}
  </div>
  <h2>Время обработки (Process)</h2><div class="col-group">
    {{range $i := .RangeT}}<label>Type {{add $i 1}}:<input type="number" name="process{{$i}}" min="0" required></label>{{end}}
  </div>
  <h2>Дедлайны для копий</h2>{{range $i := .RangeT}}
    <fieldset><legend>Type {{add $i 1}}</legend>
      <div class="col-group">
        {{range $j := $.RangeP}}<label>Copy {{add $j 1}}:<input type="number" name="deadline{{$i}}_{{$j}}" min="0" required></label>{{end}}
      </div>
    </fieldset>
  {{end}}
  <button type="submit">Вычислить</button>
</form></div></body></html>`,
	),
)

var TPLResult = template.Must(
	template.New("result").Funcs(
		template.FuncMap{
			"mul":   func(a, b int) int { return a * b },
			"add":   func(a, b int) int { return a + b },
			"plus1": func(i int) int { return i + 1 },
			"hue":   func(t, total int) int { return t * 360 / (total + 1) },
		},
	).Parse(
		`
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Результат flat-DP</title>
  <style>
    body { margin:0; padding:0; font-family:Arial,sans-serif; background:#f5f6fa }
    .container {
      max-width:900px; margin:40px auto;
      background:#fff; border-radius:8px;
      box-shadow:0 2px 8px rgba(0,0,0,0.1);
      padding:20px;
    }
    h1 { text-align:center; color:#0984e3 }
    .info { background:#dfe6e9; padding:15px; border-radius:4px; margin-bottom:20px }
    pre  { background:#f0f0f0; padding:15px; border-radius:4px; font-family: monospace; white-space:pre }
    a.button, a.back-link {
      display:inline-block; margin:10px 10px 0; padding:8px 16px;
      text-decoration:none; border-radius:4px;
    }
    a.button    { background:#0984e3; color:#fff }
    a.back-link { color:#2d3436 }
    a.button:hover    { background:#74b9ff }
    a.back-link:hover { color:#0984e3 }

    /* Gantt */
    .gantt-wrapper {
      overflow-x:auto;
      border:1px solid #ccc;
      background:#fafafa;
      padding-bottom:10px;
      margin-top:20px;
    }
    .chart {
      position:relative;
      height:40px; /* однострочная диаграмма */
    }
    .gantt-setup,
    .gantt-bar {
      position:absolute;
      height:30px;
      top: 0;
      line-height:30px;
      color:#fff;
      font-size:12px;
      white-space:nowrap;
      text-shadow:1px 1px 2px rgba(0,0,0,0.5);
      border-radius:4px;
      box-sizing: border-box;
    }
    .gantt-setup {
      opacity:0.6;
      text-align:center;
      border-radius:4px 0 0 4px;
      z-index: 1;
    }
    .gantt-bar {
      z-index: 2;
      padding: 0;
    }

    .x-axis {
      position:relative;
      margin-top:10px;
      height:20px;
      overflow: visible;
    }

    .x-axis-line {
      position:relative;
      height:100%;
      border-top:1px solid #333;
      width: {{add (mul .MinVal 10) 80}}px;
    }

    .x-axis span {
      position:absolute;
      top:0;
      font-size:12px;
      color:#333;
    }

	.params-table table {
	  width: 100%;
	  border-collapse: collapse;
	  font-family: monospace;
	  margin-top: 10px;
	}
	
	.params-table th,
	.params-table td {
	  padding: 6px 10px;
	  text-align: right;
	  border: 1px solid #ccc;
	}
  </style>
</head>
<body>
  <div class="container">
    <h1>Результаты</h1>
    <div class="info">
      <p><strong>T = {{.T}}</strong>, <strong>P = {{.P}}</strong></p>
      <p>Минимальное время завершения: <strong>{{.MinVal}}</strong></p>
      <p>Тип последней задачи: <strong>{{.BestLast}}</strong></p>
      <p>Время выполнения DP: <strong>{{printf "%.6f" .DPTime}} сек</strong></p>
    </div>
    <h2>Параметры задачи:</h2>
		<div class="params-table">
		  <table>
		    {{range $i, $row := .ParamsTable}}
		      <tr>
		        {{range $cell := $row}}
		          {{if eq $i 0}}
		            <th>{{$cell}}</th>
		          {{else}}
		            <td>{{$cell}}</td>
		          {{end}}
		        {{end}}
		      </tr>
		    {{end}}
		  </table>
		</div>
    <h2>Диаграмма Ганта</h2>
    <div class="gantt-wrapper">
      <div class="chart">
        {{range $s := .Steps}}
          {{if gt $s.SetupLen 0}}
            <div class="gantt-setup"
                 style="
                   left: {{add (mul $s.Start 10) 40}}px;
                   width: {{mul $s.SetupLen 10}}px;
                   background: hsl({{hue $s.Type $.T}},20%,40%);
                 ">
              S{{$s.Type}}
            </div>
          {{end}}
          <div class="gantt-bar"
               style="
                 left: {{add (mul (add $s.Start $s.SetupLen) 10) 40}}px;
                 width: {{mul $s.ProcLen 10}}px;
                 background: hsl({{hue $s.Type $.T}},70%,50%);
               "
               title="Тип {{$s.Type}}, копия {{$s.Instance}}">
            T{{$s.Type}}#{{$s.Instance}}
          </div>
        {{end}}
      </div>

      <!-- Ось X -->
      <div class="x-axis">
        <div class="x-axis-line">
          {{range $tick := .Ticks}}
            <span style="left:{{add (mul $tick 10) 40}}px">{{$tick}}</span>
          {{end}}
        </div>
      </div>
    </div>

    <a class="button" href="/download">Скачать input.gms</a>
    <a class="back-link" href="/">← Вернуться к началу</a>
  </div>
</body>
</html>
`,
	),
)

var TPLError = template.Must(
	template.New("error").Parse(
		`<!DOCTYPE html>
<html lang="ru"><head><meta charset="UTF-8"><title>Ошибка</title>
<style>body{font-family:Arial,sans-serif;background:#ffe6e6;margin:0;padding:0}
.container{max-width:500px;margin:80px auto;background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1);padding:30px;text-align:center}
h1{color:#c0392b}p{margin:20px 0}a.button{display:inline-block;padding:10px 20px;background:#c0392b;color:#fff;text-decoration:none;border-radius:4px}
a.button:hover{background:#e74c3c}</style>
</head><body><div class="container">
<h1>Ошибка: расписание не найдено</h1>
<p>Невозможно уложиться в заданные дедлайны.</p>
<a class="button" href="/">Вернуться назад</a>
</div></body></html>`,
	),
)
