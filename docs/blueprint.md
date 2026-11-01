# Delta: la parte determinista

*Versión 1, 1 de noviembre de 2026. Diseño del CLI `delta` en Go: estructura, capas, paquetes, librerías y reglas de código.*

Este documento explica cómo se construye la parte determinista del harness: todo lo que un programa puede decidir sin un modelo. Complementa a `direction.md` (qué es Delta y por qué), `backstory.md` (contexto) y `waypoints.md` (en qué orden). Si algo choca con `direction.md`, manda `direction.md`.

**Cómo leerlo:** lo marcado **Decidido** salió de la conversación con Roney y se da por cerrado. Lo marcado *propuesta* es mi criterio y se puede cambiar. Lo marcado *a verificar* es algo que no comprobé y hay que confirmar antes de implementar. Ningún código de este documento existe todavía: los fragmentos de Go son bocetos para fijar los nombres y las formas, no una implementación.

---

## 1. Qué es y qué no es

La idea central de Delta es **el modelo propone, el código decide** (`direction.md`, §1). Esta parte es el "código decide":

- Valida que un spec tenga la forma correcta, con IDs únicos y tags válidos.
- Ejecuta los comandos de verificación de cada claim y dice `Holds`, `Violated` o `Unverified`.
- Lleva el estado de cada slice (`draft`, `approved`, `built`, `verified`, `stale`).
- Detecta que el código cambió y el contrato quedó viejo.
- Valida los mensajes de commit y mide el presupuesto de contexto.
- Decide, para los hooks de Claude Code, qué se permite y qué se bloquea.
- Instala el harness (skills, rules, hooks) en un proyecto.

Lo que **no** hace: no llama a ningún modelo, no entrevista, no redacta claims, no escribe código del proyecto. Eso es de Quill, Forge y Seal.

El CLI **no depende de Claude Code**. Los hooks son un adaptador fino encima (sección 4). Otros agentes no son una meta, pero la puerta queda abierta sin esfuerzo.

---

## 2. Decisiones ya tomadas

| # | Decisión | Estado |
|---|---|---|
| 1 | Lenguaje: Go, un binario sin runtime | Decidido |
| 2 | Dos carpetas de código: `scripts` y `source`; el harness vive en `source/harness` | Decidido |
| 3 | `go.mod` en la raíz del repo | Decidido |
| 4 | Cuatro capas dentro de `source`: `domain`, `usecase`, `gateway`, `adapter` | Decidido |
| 5 | Nombres de paquete de una palabra, de longitud pareja: 7 a 9 letras en `domain`, 5 a 6 en `usecase` y `gateway`, y `cmd` y `hook` en `adapter` | Decidido |
| 6 | CLI con `spf13/cobra` | Decidido |
| 7 | Se añade el comando `delta hook`, sin argumento: el evento llega en el JSON | Decidido |
| 8 | Los comandos de verificación se ejecutan directo con `os/exec`, sin shell | Decidido (`direction.md`, 6.4) |
| 9 | `delta check` no lee la salida de los runners: `Holds` exige exit 0 más un test que cite el ID | Decidido (`direction.md`, 6.4 y 6.5) |
| 10 | Delta no se usa sobre sí mismo por ahora: este repo no tiene `.claude/`, `deltas/`, `deltas.config` ni hook de git. `source/harness/` es la fuente de lo que `delta init` instala en otros proyectos | Decidido, hecho |
| 11 | Frontmatter plano (`clave: valor`) con parser propio; parser de claims propio | Propuesta |
| 12 | Salvo cobra, solo la librería estándar (sección 9) | Propuesta |
| 13 | Archivos de código de `domain` y de `adapter/cmd` de 6 a 7 letras; su prueba se llama `<nombre>_test.go` | Decidido |
| 14 | Los linters de lenguaje los maneja cada proyecto. `PostToolUse` no corre ninguno: corre `delta lint` cuando Claude escribe un archivo de test, y le cuenta solo los problemas de ese archivo | Decidido, hecho (sección 8.5) |
| 15 | El binario conserva el nombre `delta`; el hook `commit-msg` comprueba la primera línea de `delta version` antes de ejecutarlo | Decidido, hecho |
| 16 | `Stop` exige un `delta check` sobre lo que cada slice mira del árbol (todo el código, salvo lo que `deltas.config` acote); la marca vive en la caché del usuario, no en el proyecto ni en `.git` | Decidido, hecho (sección 8.7) |
| 17 | `deltas/deltas.config` (antes `deltas.config` en la raíz), claves planas, que absorbe a `budget.conf` y trae los patrones de archivos de test (sección 8.8) | Decidido, hecho |
| 18 | ~~Varios `deltas/` por proyecto~~ Retirado: un solo `deltas/` en la raíz, también en un monorepo, para que la KB no se disperse. Un slice es un nombre sin carpeta y sus comandos corren desde la raíz | Decidido, hecho |
| 19 | `deltas/`, `specifies/` y `.spec.md` son el formato del producto y quedan fijos; lo que ata a un lenguaje (qué es un archivo de test) se configura | Decidido |
| 20 | Todo lo de Delta que no es de Claude Code vive en `deltas/`: `deltas.config`, `deltas.manifest` (reservado, ver abajo) y `specifies/` con los specs. La raíz del proyecto solo gana el `CLAUDE.md`. Hay un único `deltas/`, el de la raíz (decisión 18) | Decidido, hecho (el manifest, reservado) |
| 20 | Los hooks de `settings.json` tienen un límite de 20 s | Decidido, hecho |

---

## 3. Estructura del repositorio

```
deltas-ai/
├── go.mod
├── scripts/
│   └── delta/main.go            # raíz de composición: crea las piezas y las conecta
├── source/
│   ├── domain/                  # puro: sin I/O, solo librería estándar
│   │   ├── findings/            # Diagnostic: archivo, línea, código estable y mensaje
│   │   ├── assertion/           # Claim, ID, Tag, enunciado EARS y Given/When/Then
│   │   ├── specfile/            # Spec, Frontmatter, versión de formato, parser y render
│   │   ├── lifecycle/           # Slice, Status, transiciones válidas
│   │   ├── judgment/            # Verdict (Holds, Violated, Unverified), la regla de evidencia y Mark/Standing
│   │   ├── drift/               # regla de stale: archivos cambiados vs referenciados
│   │   ├── linkcheck/           # reglas de coherencia entre specs y tests
│   │   ├── commits/             # validador de Conventional Commits
│   │   ├── knowledge/           # qué archivos son de la base de conocimiento
│   │   ├── budget/              # pesa lo siempre cargado contra sus techos
│   │   └── conffile/            # deltas.config: lo que un proyecto configura
│   ├── usecase/                 # orquestación; declara los puertos (interfaces) que necesita
│   │   ├── setup/               # delta init
│   │   ├── author/              # spec new, spec write, link, approve
│   │   ├── verify/              # lint, check
│   │   ├── status/              # status, stale
│   │   ├── guard/               # decisiones de hooks: qué se permite, qué se bloquea
│   │   └── measure/             # pesa lo siempre cargado: delta budget
│   ├── gateway/                 # los únicos efectos: implementan los puertos
│   │   ├── vaults/              # lectura y escritura atómica de la KB y de los archivos del proyecto
│   │   ├── runner/              # os/exec sin shell, con timeout
│   │   ├── gitcli/              # envoltorio fino del binario git, y la huella del árbol de trabajo
│   │   ├── ledger/              # qué árbol revisó cada delta check (en la caché del usuario)
│   │   └── bundle/              # el harness embebido, listo para instalar
│   ├── adapter/                 # traducen el mundo exterior a casos de uso
│   │   ├── cmd/                 # comandos cobra y formato de salida
│   │   └── hook/                # JSON de Claude Code por stdin → exit code
│   └── harness/                 # lo que `delta init` instala (se embebe en el binario)
│       ├── harness.go           # //go:embed
│       ├── core/                # lo que se instala tal cual, sin ser skill ni regla
│       │   ├── configs/         # settings.json (hooks y permisos) y budget.conf
│       │   ├── githooks/        # commit-msg (se instala como .githooks/)
│       │   └── templates/       # spec.md, CLAUDE.md
│       ├── skills/              # quill, forge, seal
│       └── rules/               # core (R1, R2) y base (B1, B2)
```

`source/harness/core/` y `source/harness/rules/core/` no son lo mismo: la primera agrupa los archivos de soporte que `delta init` copia al proyecto; la segunda son las dos reglas siempre cargadas (R1 y R2). Git no versiona carpetas vacías, así que `githooks/` y `templates/` se crean cuando tengan contenido (`templates/` en la Fase 1, con `delta spec new`; `githooks/` en la Fase 2).

**Por qué `scripts` y no `cmd`:** es una preferencia de nombre. `go install ./scripts/delta` sigue produciendo un binario llamado `delta`, porque el nombre sale de la carpeta que contiene el `main`.

**Por qué `source`:** es una preferencia de nombre. Al no llamarse `internal`, el compilador ya no impide que otros módulos importen estos paquetes; la API pública de Delta es el CLI y el formato del spec, no los paquetes Go.

**Lo que no está en esta estructura:** `docs/` (documentos de trabajo) sigue en la raíz. Este repo no tiene `.claude/`, `deltas/`, `deltas.config` ni `.githooks/`: Delta no se construye con Delta por ahora (decisión 10); el primer uso real es otro proyecto (Fase 6).

---

## 4. Las capas y la regla de dependencia

```
adapter  ──>  usecase  ──>  domain
                 ^
              gateway     (implementa lo que usecase declara)
```

| Capa | Qué contiene | Qué puede importar |
|---|---|---|
| `domain` | Las reglas del negocio: qué es un claim, cuándo una transición es válida, qué significa `Holds`. Funciones puras | Solo la librería estándar y otros paquetes de `domain` |
| `usecase` | Los pasos de cada comando: cargar, decidir, guardar. Declara las interfaces (puertos) que necesita | `domain` |
| `gateway` | La única capa con efectos: disco, procesos, git, el harness embebido. Implementa los puertos | `usecase` (para las interfaces) y `domain` |
| `adapter` | La entrada y la salida: argumentos de línea de comandos, JSON de hooks, texto de pantalla, exit codes | `usecase` y `domain` |

La raíz de composición, `scripts/delta/main.go`, es el único lugar que conoce todas las capas. Crea los gateways, se los pasa a los casos de uso por constructor, y los casos de uso a los adaptadores. No hay variables globales ni `init()`.

**Cómo se hace cumplir.** Un test lee los imports de cada paquete con `go/parser` y falla si `domain` importa algo fuera de la librería estándar y de `domain`, si `usecase` importa `gateway` o `adapter`, o si `adapter` importa `gateway`. Es una regla detectable con código, así que se detecta con código (pilar 9). Es el primer test de la Fase 1.

**Los puertos viven donde se usan.** Cada caso de uso declara interfaces de una o dos funciones con lo que necesita. Boceto:

```go
// usecase/verify
type Specs interface {
    Load(ctx context.Context, slice string) ([]byte, error)
    List(ctx context.Context) ([]string, error)
}

type Runner interface {
    // Devuelve solo el código de salida. La salida del comando se muestra, no se interpreta.
    Run(ctx context.Context, command string) (exit int, err error)
}
```

---

## 5. Los paquetes, uno por uno

### 5.1 `domain`

| Paquete | Responsabilidad | Tipos y funciones principales |
|---|---|---|
| `findings` | El hallazgo que todo validador reporta | `Diagnostic` (archivo, línea, código estable, mensaje) |
| `assertion` | El átomo del contrato | `ID`, `ParseID(string) (ID, error)`, `Kind` (`Invariant`, `Acceptance`), `Method` (`Test`, `Manual`, `None`: cómo se verifica un claim), `Tag` (método y argumento), `Claim` (ID, tag, texto, línea), `CheckStatement(kind, texto) error` |
| `specfile` | El archivo completo | `Spec` (formato, slice, estado, claims), `Parse(name, src) (Spec, []Diagnostic)`, `SetStatus(src, status) ([]byte, error)`, `SetTag(src, id, tag) ([]byte, error)` |
| `lifecycle` | El ciclo de vida | `Status` (`Draft`, `Approved`, `Built`, `Verified`, `Stale`), `Can(from, to) bool`: el ciclo, y la vuelta a `draft` desde cualquier otro estado |
| `judgment` | El resultado de verificar | `Verdict` (`Holds`, `Violated`, `Unverified`), `Outcome` (claim, veredicto, motivo), `Judge(tag, exit, cited) Outcome`; `Mark` (el árbol que vio un `delta check` y si hubo violación) y `Mark.At(tree)`, que dice si es `Unchecked`, `Failing` o `Passing` para el árbol de ahora |
| `drift` | Cuándo un contrato está viejo | `Stale(referenced, changed []string) bool` y qué archivos coinciden |
| `linkcheck` | Coherencia entre specs y tests | Reglas: ID sin test que lo cite, test que cita un ID inexistente, claim sin tag, `skip`/`only` en archivos que citan IDs. `Scope` es la parte del proyecto de la que depende un slice: todos los archivos de código, o solo las rutas que `deltas.config` declare para él, más su propio spec; el Markdown no cuenta. `TestMatcher` dice qué archivos son tests, con los patrones de `deltas.config` o con la regla propia de Delta. `Under(files, dir)` nombra los archivos desde la carpeta en que corre un comando |
| `commits` | Conventional Commits | `Validate(message) []Diagnostic`: tipos permitidos, longitud de 40 a 50, imperativo, minúscula, sin punto final |
| `knowledge` | Qué archivos son de la KB | `Covers(path) bool`: la raíz de `CLAUDE.md`, `.claude/{docs,rules,agents,skills}/` y, en cualquier carpeta `deltas/` a cualquier profundidad, sus archivos directos (`deltas.config`, `deltas.manifest`) y los de su `specifies/` |
| `budget` | Presupuesto de contexto | `Measure(files, ceilings) Report` pesa el router, las reglas sin `paths:` y las descripciones de skills y agentes; `Report.Findings()` nombra cada techo superado; `ParseCeilings` lee `budget.conf` |
| `conffile` | Lo que un proyecto configura | `Parse(text) Conf`: los techos del presupuesto, `test_files`, `command_timeout`, `required_tools` y `scope.<slice>`; las líneas que no sirven se dejan fuera y se cuentan en `Conf.Problems` (sección 8.8) |

Observaciones:

- **`specfile` edita por líneas.** Cambiar el estado o el tag de un claim modifica únicamente esa línea y deja el resto del archivo intacto. Es lo que permite que `delta link` y `delta approve` toquen un solo renglón.
- **`Judge` es la regla de evidencia en una función:** exit distinto de 0 es `Violated`; exit 0 con un test que cita el ID es `Holds`; exit 0 sin test que cite el ID es `Unverified` con el motivo escrito; `manual` y `none` son `Unverified`. Es la pieza más fácil de probar de todo el proyecto.
- **`linkcheck` reconoce el ID con coincidencia tolerante** (`INV-01`, `INV_01`, `INV01`, sin distinguir mayúsculas, con un límite al final para que `INV01` no coincida con `INV011`), porque en Go los nombres de función no admiten guiones.
- **`linkcheck.Cited` acota la cita a las rutas del comando.** Los IDs son únicos solo por spec, así que un test de `facturacion` no puede dar `Holds` al `INV-01` de `pagos`. Una ruta es un argumento que empieza por `./`, `../` o es `.`, o que lleva una barra y termina en un archivo de script; los flags y sus valores no cuentan. *Shortcut: no se interpretan comillas ni rutas con espacios; si ningún argumento parece ruta, vale cualquier test.*
- **`linkcheck.Check` emite avisos además de errores.** Un claim `test` que nadie cita es un error (`LINT-NO-CITE`). Un test que cita un ID que ningún spec define (`LINT-UNKNOWN-ID`, solo en mayúsculas para no confundirlo con una variable `inv1`) y un archivo que cita IDs y salta o enfoca un test (`LINT-SKIP`) son avisos (`Diagnostic.Warn`): se imprimen pero no fallan la ejecución.
- **`commits.Validate` ignora las líneas `#`** (git las quita), deja pasar los mensajes que git genera (`Merge `, `Revert "`, `fixup! `, `squash! `) y no comprueba el modo imperativo, porque ninguna regla distingue "embed" de "embedded" sin un diccionario.

### 5.2 `usecase`

| Paquete | Comandos | Qué orquesta |
|---|---|---|
| `setup` | `delta init`, y el aviso de `status` y `SessionStart` | `Run` escribe el esqueleto, instala el harness desde `bundle` y fija `core.hooksPath` sin pisar uno ajeno; es idempotente. `Check` lista los problemas de instalación, sin fallar. `Check` incluye también los problemas de `deltas.config` y las herramientas que pide y no están en el `PATH` |
| `author` | `delta spec new`, `spec write`, `link`, `approve`, `built` | Valida con `domain` y escribe con `vaults`. Es el único camino de escritura de la KB, y aplica las reglas de estado de `spec write` (sección 7.2) |
| `verify` | `delta spec lint`, `lint`, `check` | Carga specs, busca tests que citen IDs, ejecuta comandos con `runner`, aplica `judgment.Judge` y anota qué slices revisó y si hubo violación (`ledger`) |
| `status` | `delta status`, `ready`, `stale` | Reúne estado, cobertura y deriva por slice; compara con `git`. `Ready` dice si un slice puede construirse |
| `guard` | `delta hook` | Responde "¿se permite esto?" para cada evento: editar la KB, commitear, empezar la sesión y terminar el turno (`Stop`) |
| `measure` | `delta budget` | Lee los archivos que se cargan siempre y los pesa contra los techos que le da `deltas.config` (o los valores por defecto) |

### 5.3 `gateway`

| Paquete | Efecto | Detalles |
|---|---|---|
| `vaults` | Lee y escribe la KB y los archivos del proyecto | Escritura atómica: archivo temporal en la misma carpeta y renombrado. Conserva LF. Localiza la raíz del repo y lista y lee los specs de `deltas/specifies/` (`List`, `Load`). Ningún camino sale del proyecto |
| `runner` | Ejecuta comandos | `os/exec` con `context` y timeout, desde la carpeta que se le indique. Parte el comando en argumentos sin expandir variables. *Shortcut: sin shell, no admite pipes ni redirecciones; si hacen falta, se escribe un programa* |
| `gitcli` | Llama a `git` | Leer y fijar `core.hooksPath`, y `Tree(scope)`: la huella de la parte del árbol de trabajo que cubre un `Scope` (sección 8.7). Se usa el binario de git, no una librería |
| `ledger` | Recuerda qué árbol revisó cada `delta check` | Un archivo JSON por proyecto en `os.UserCacheDir()/delta/`, con la ruta de la raíz como clave. Si falta o se daña, `Stop` pide un nuevo `delta check`; nunca falla un comando |
| `bundle` | Entrega el harness | Envuelve el `embed.FS` de `source/harness/` y sabe a qué ruta del proyecto va cada archivo |

### 5.4 `adapter`

| Paquete | Entrada | Salida |
|---|---|---|
| `cmd` | Argumentos y flags (cobra) | Texto en tabla con `text/tabwriter`, o JSON con `--json`; exit code |
| `hook` | JSON del evento de Claude Code por stdin | Exit 0 para permitir, exit **2** para bloquear (el 1 no bloquea), con el motivo en stderr |

`hook` no contiene reglas. Extrae del JSON lo que `guard` necesita (herramienta, ruta, comando), llama a `guard` y traduce la decisión a un exit code.

---

## 6. De los comandos a los paquetes

| Comando | Adaptador | Caso de uso | Dominio que decide |
|---|---|---|---|
| `delta init` | `cmd` | `setup` | `lifecycle`, `specfile` (plantilla) |
| `delta spec new <slice>` | `cmd` | `author` | `specfile`, `lifecycle` |
| `delta spec lint [slice]` | `cmd` | `verify` | `specfile`, `assertion` |
| `delta spec write` | `cmd` | `author` | `specfile` |
| `delta link <slice> <claim-id> <comando>` | `cmd` | `author` | `specfile`, `assertion` |
| `delta lint` | `cmd` | `verify` | `linkcheck`, `assertion` |
| `delta check [slice]` | `cmd` | `verify` | `judgment`, `linkcheck` |
| `delta status` | `cmd` | `status` | `lifecycle`, `judgment`, `drift` |
| `delta stale` | `cmd` | `status` | `drift` |
| `delta budget` | `cmd` | `measure` | `budget` |
| `delta version` | `cmd` | (ninguno) | (ninguno) |
| `delta commit-msg <archivo>` | `cmd` | (ninguno: `cmd` llama a `domain/commits`; `guard` nace en la Fase 2, cuando decida algo propio) | `commits` |
| `delta approve <slice>` | `cmd` | `author` | `lifecycle` |
| `delta built <slice>` | `cmd` | `author` | `lifecycle` |
| `delta ready <slice>` | `cmd` | `status` | `lifecycle` |
| `delta context <slice>` | `cmd` | `status` | `specfile`, `assertion` |
| `delta hook` | `hook` | `guard` | `knowledge`, `commits` |

`delta init` instala, sin pisar nada que el proyecto ya tenga: `.githooks/commit-msg` y `core.hooksPath`, los skills y las reglas bajo `.claude/`, `deltas.config`, un `CLAUDE.md` mínimo con el nombre del proyecto, la carpeta `deltas/specifies/` y los hooks y permisos de Delta fusionados en `.claude/settings.json` (sección 12). `delta hook` se registra en `source/harness/core/configs/settings.json` y atiende todos los eventos: el evento viene en el JSON.

---

## 7. Contratos de datos

### 7.1 El spec

Definido en `direction.md`, 6.4. Resumen de lo que el parser exige:

- Frontmatter con `format`, `slice` y `status`, claves planas. Una clave desconocida, repetida o ausente es un error; `slice` va en minúsculas unidas por guiones.
- Cada claim es una línea `- <ID> [<tag>]` seguida de su texto indentado, que puede ocupar varias líneas (se une con espacios).
- IDs únicos dentro del spec, con dos prefijos: `INV-<n>` (invariante) y `ACC-<n>` (aceptación). El prefijo decide la notación de la frase.
- Tags: `test: <comando>`, `manual` (con un paso opcional, `manual: <paso>`) y `none`. El tag `check` de la generación anterior ya no existe.
- `format: 1` permite que una versión futura migre specs viejos.
- Un `INV` se escribe en **EARS**: el parser valida la estructura, no un verbo modal (la generación anterior escribe sin "shall"). Una cláusula `When`, `While`, `Where` o `If` lleva una coma antes de la respuesta, y `If` exige `then` después de la coma; se pueden encadenar. Un `ACC` se escribe `Given <contexto>, when <acción>, then <resultado>`. Las palabras clave se aceptan en inglés y en español (`cuando`, `mientras`, `donde`, `si`/`entonces`, `dado que`), y una misma frase puede mezclarlas.
- Se ignoran el código entre vallas (` ``` `) y los comentarios HTML, para poder mostrar un claim de ejemplo sin declararlo. Un ítem de lista cuya primera palabra parece un ID (`ISO-8601`) se toma como intento de claim y se reporta.
- `Parse` devuelve todos los diagnósticos, no solo el primero, y en `Claims` solo los claims que se leyeron sin problemas. Los códigos (`SPEC-FRONTMATTER`, `SPEC-KEY`, `SPEC-FORMAT`, `SPEC-SLICE`, `SPEC-STATUS`, `SPEC-CLAIM`, `SPEC-TAG`, `SPEC-DUP-ID`, `SPEC-NO-TEXT`, `SPEC-STATEMENT`) son estables.

### 7.2 Estados del slice

```
draft  ->  approved  ->  built  ->  verified
                           ^           |
                           +-- stale <-+
```

`lifecycle.Can(from, to)` es la única fuente de qué transiciones son legales. Solo el CLI modifica el estado, y cada cambio edita únicamente la línea `status:` (`specfile.SetStatus`).

| Transición | Quién la dispara |
|---|---|
| `draft → approved` | `delta approve <slice>`, una persona. Exige un spec sin problemas y con claims; `settings.json` obliga a Claude Code a pedir confirmación a la persona |
| `approved → built` y `stale → built` | `delta built <slice>`, que ejecuta Forge al cerrar su trabajo |
| `built → verified` | `delta check`, cuando todos los claims `test` del slice se cumplen y hay al menos uno; los `manual` y `none` quedan visibles como deuda. *Se pide uno como mínimo para que un slice sin nada ejecutado no figure como verificado* |
| `verified → stale` | `delta stale`, en la Fase 5 |
| cualquiera `→ draft` | `spec write`, cuando cambia el contenido de un spec que ya pasó de `draft` (aviso `AUTHOR-REAPPROVE`): un contrato distinto necesita otra aprobación |

`spec write` no puede mover un spec hacia delante: uno nuevo solo puede escribirse como `draft` (`AUTHOR-NEW-STATUS`) y uno existente no puede declarar un estado que no tiene (`AUTHOR-SET-STATUS`). Cambiar solo los tags con `delta link` no cambia el estado.

`delta ready <slice>` es la compuerta de Forge: sale con 0 solo si el spec no tiene problemas y está `approved` o `stale`; si no, sale con 1 y dice qué falta.

### 7.3 Diagnósticos y códigos de salida

- Un **diagnóstico** lleva archivo, línea, un código estable por tipo (por ejemplo `SPEC-DUP-ID`) y un mensaje. Hooks y CI pueden depender del código, no del texto.
- **Códigos de salida del CLI:** 0 todo bien, 1 se encontraron violaciones, 2 uso incorrecto.
- **En `delta hook`:** 0 permite y 2 bloquea, porque eso es lo que Claude Code espera.
- **Salidas del CLI:** los hallazgos (diagnósticos) van a stderr y los resultados (tablas) a stdout; la salida de los comandos de verificación va a stderr. Los avisos (`Warn`) se imprimen pero no cambian el código de salida.
- Con `--json` la salida es estable y ordenada, pensada para hooks y CI.

---

## 8. Las piezas críticas, paso a paso

### 8.1 El parser de specs

Es la pieza de la que depende todo lo demás, así que se escribe primero. Trabaja línea por línea con `bufio` y `regexp`; no usa un parser de Markdown general porque el formato es un subconjunto fijo y un parser general escondería los números de línea de los errores. Devuelve el `Spec` y la lista completa de diagnósticos, no solo el primero. Tiene fuzzing nativo de Go, porque es lo que más datos ajenos recibe.

### 8.2 `delta check <slice>`

1. `verify` carga el spec y lo parsea. Si hay diagnósticos de error, termina con ellos.
2. Para cada claim con tag `test`, busca en los archivos de tests que el comando ejecuta (por las rutas que nombra) si alguno cita el ID (`linkcheck.Cited`).
3. `runner` ejecuta el comando desde la raíz del repo y devuelve el código de salida. Las rutas del comando son relativas a ella. La salida del comando se muestra a la persona pero **no se interpreta**.
4. `judgment.Judge` combina tag, código de salida y "¿algún test cita el ID?" en un `Outcome`. Un comando que no pudo ejecutarse (no existe, o pasó el tiempo límite) deja el claim `Unverified` con el motivo: no dice nada sobre el claim.
5. Los claims `manual` y `none` se reportan como `Unverified`.
6. Se anota, por cada slice revisado, la huella del árbol y si hubo algún `Violated` (sección 8.7). Si no se puede anotar, es un aviso (`CHECK-RECORD`) y el resultado no cambia.
7. Se imprime la tabla y se sale con 1 si hay algún `Violated`.

Qué garantiza y qué no: un `Holds` prueba que el comando salió con 0 y que existe un test que cita el ID. **No prueba que ese test se ejecutó** (un `skip` pasaría, y un comando que no nombra rutas acepta cualquier test). Ese límite está documentado en `direction.md`, §10, junto con la mejora futura (un archivo de evidencia que cada test escribe).

### 8.3 `delta stale`

Para cada spec, `drift` compara los archivos que el spec referencia con los que `gitcli` dice que cambiaron desde la última verificación. La granularidad es el archivo (decisión de `direction.md`). Marca de más si el cambio no altera el comportamiento, y no marca si el comportamiento cambia sin tocar los archivos nombrados. Está documentado como heurística.

### 8.4 Escritura atómica de la KB

`author` valida con `domain`, y solo si no hay errores `vaults` escribe en un archivo temporal de la misma carpeta y lo renombra sobre el destino. Un spec nunca queda a medias.

### 8.5 `delta hook`

Un solo comando atiende todos los eventos: `hook` lee el JSON de stdin, `guard` decide y `hook` traduce la decisión a un código de salida. `settings.json` registra el mismo `delta hook` para todos.

**El contrato de Claude Code**, verificado el 1 de noviembre de 2026 en https://code.claude.com/docs/en/hooks:
- **Configuración:** `hooks.<Evento>` es una lista de grupos `{ matcher, hooks: [{ type: "command", command, timeout }] }`. El `matcher` es una lista exacta separada por `|` (`Write|Edit`); cualquier otro carácter lo convierte en una expresión regular.
- **Entrada por stdin:** `hook_event_name`, `cwd`, `session_id`, `transcript_path` y `permission_mode`, más `tool_name` y `tool_input` en los eventos de herramienta (`command` en Bash; `file_path` absoluto en Write y Edit). `PostToolUse` añade `tool_response`; `Stop`, `stop_hook_active` y `last_assistant_message`; `SessionStart`, `source`.
- **Salida:** exit **2** bloquea (en `PreToolUse` la acción no ocurre y el stderr llega a Claude; en `Stop` impide parar). Cualquier otro código, incluido el 1 y el 127 de "comando no encontrado", **no bloquea**. En `SessionStart`, el stdout con exit 0 se suma al contexto de Claude.
- **`PostToolUse`:** el exit 2 no bloquea nada (la herramienta ya corrió) pero el stderr llega a Claude; con exit 0 Claude no ve ni el stdout ni el stderr. Verificado el 1 de noviembre de 2026 en la misma página.
- **`Stop`:** `stop_hook_active` es verdadero cuando Claude ya continúa por un hook de parada; hay que mirarlo para no bloquear algo que nunca se resolverá, y Claude Code corta tras 8 continuaciones seguidas.

| Evento | Qué hace hoy |
|---|---|
| `SessionStart` | Imprime el resumen de los slices (estado y claims por tag), los problemas de instalación (`setup.Check`) y el hecho de que la KB no se edita directo |
| `PreToolUse` con `Write` o `Edit` | Bloquea (2) si el archivo es de la KB (`knowledge.Covers`); una persona sí puede editarla a mano |
| `PreToolUse` con `Bash` | Si la línea es un `git commit` con un mensaje legible (`commits.FromCommand`), lo valida y bloquea (2) si no cumple |
| `PostToolUse` con `Write` o `Edit` | Si el archivo es de test (`linkcheck.IsTestFile`), corre `delta lint` y le cuenta a Claude (exit 2: la herramienta ya corrió, pero Claude sí ve el stderr) los problemas de **ese** archivo. Los de otros archivos no se dicen: un claim sin test aún es normal mientras se escriben los tests |
| `Stop` | Bloquea (2), una sola vez, si hay cambios en lo que un slice con claims `test` mira del árbol y no hay un `delta check` sobre ellos, o lo hay con violaciones (sección 8.7). Si el slice está `approved` o `stale` y su alcance cambió, el mensaje es otro: `/forge <slice>`, que termina con `delta built`. Un slice aprobado cuyo código no ha cambiado no es trabajo pendiente |

**Decisiones y límites:**
- Todos los hooks de `settings.json` llevan `timeout: 20` (segundos): `PostToolUse` lee todos los archivos de test y `Stop` hace una pasada de git por slice. El valor por defecto de Claude Code sería de 600.
- Un fallo propio (JSON ilegible, error de disco) sale con **1**, nunca con 2: un delta roto no debe bloquear al agente. Si `delta` no está instalado, el shell sale con 127, que tampoco bloquea.
- `Write` y `Edit` son las únicas herramientas de archivo cubiertas, porque son las dos cuyo `file_path` confirmé. `MultiEdit` y `NotebookEdit` no se cubren hasta ver sus campos.
- *Shortcut: el hook ve la línea de comando, no sus efectos; un `sed -i` o un `cat >` sobre un spec desde Bash no se detecta.*
- El mensaje de commit se lee de `-m` y `--message` (también varios) y de la forma `$(cat <<'EOF' ... EOF)`. Con `-F`, con `--amend` sin `-m` o con otra sustitución de shell no se puede leer y se deja pasar: lo cubren el git hook y CI.
- `settings.json` no usa el filtro `if` de los hooks (`Bash(git commit*)`): no pude confirmar cómo evalúa comandos compuestos, y si no coincide el commit no se vigila. Se paga un proceso `delta` por cada llamada a Bash.

### 8.6 `delta budget`

Pesa lo que se carga en **cada turno** y falla (exit 1) si se supera algún techo; es el gate de la sección 6.7. En CI lo cubre una prueba que instala el harness en un repositorio temporal y lo mide. Lo siempre cargado es el router (`CLAUDE.md`), las reglas sin `paths:` y la descripción de cada skill y de cada agente. El cuerpo de un skill y las reglas con `paths:` no cuestan nada hasta que se usan.

| Clave de `deltas.config` (por defecto) | Techo de |
|---|---|
| `always_tokens` (3500) | el total de lo siempre cargado |
| `claude_lines` (80) | líneas de `CLAUDE.md` |
| `rule_always_lines` (50) | líneas de una regla sin `paths:` |
| `skill_desc_chars`, `agent_desc_chars` (1024) | caracteres de una descripción |

- Los techos salen de `deltas.config` (sección 8.8). Sin él, `delta budget` lee el `.claude/budget.conf` de la generación anterior, y sin ninguno de los dos usa esos valores por defecto. Una clave en 0 apaga su techo.
- Los tokens se estiman a **3,3 caracteres por token**, la razón con que se calibraron los techos contra `/context`, y no a 4 como decía la nota de la sección 6.1 de `direction.md`.
- Es el port de la parte "siempre cargada" de `budget.sh`, que ya no se distribuye. *Shortcut: no se miden las clases `scoped`, `invoke` y `demand`, los punteros rotos ni los globs muertos; sus claves siguen siendo válidas en `deltas.config`.*
- Sobre este repositorio coincide con el script al carácter, salvo en dos descripciones con rayas largas (`—`): el script las cuenta en bytes porque su `awk` corre en la configuración regional C, y `delta budget` cuenta caracteres, que es lo que dice el nombre de la clave.

### 8.7 La marca de `delta check` y el hook `Stop`

`Stop` quiere saber si el agente corrió `delta check` sobre lo que cambió. Cada hook es un proceso nuevo sin memoria, así que `delta check` deja una marca y `Stop` la lee.

1. **Alcance de un slice** (`linkcheck.Scope`). Lo que el slice mira. **Por defecto es todo el código del proyecto**: un comando como `go test ./billing` no dice qué código alcanza (un paquete compartido, por ejemplo), así que un cambio en cualquier parte puede romper un claim. Si `deltas.config` declara `scope.<slice>=<rutas>`, el alcance de ese slice se reduce a esas rutas. En ambos casos entra su propio spec, y el Markdown no cuenta (documentación no cambia si el código cumple el contrato).
2. **Huella** (`gitcli.Tree(scope)`). Un hash de `HEAD` más el contenido de cada archivo del alcance que difiere de `HEAD` (incluido lo preparado y lo borrado) o que no está seguido ni ignorado. No sabe de lenguajes, y vuelve al mismo valor si se revierte una edición. Si nada del alcance cambió (por ejemplo tras un commit) no hay huella y no hay nada que probar.
3. **Marca** (`ledger.Record`). Al terminar, `delta check` anota por slice revisado la huella de su alcance y si hubo un `Violated`. Se calcula después de correr los comandos, para incluir lo que ellos generen.
4. **Decisión** (`guard`, evento `Stop`). Se mira solo a los slices con claims `test` y sin problemas en su spec. Para cada uno, `ledger.Standings` dice: `Passing` (nada de su alcance cambió, o una marca lo cubre), `Unchecked` (pide `delta check`) o `Failing` (pide arreglar o avisar a la persona).
5. **Una sola vez.** Si `stop_hook_active` es verdadero, deja terminar sin mirar nada: así no hay bucles ni se llega al tope de 8 continuaciones. Un fallo propio sale con 1 y no bloquea.

**Dónde vive la marca.** En `os.UserCacheDir()/delta/<hash de la ruta del repo>.json`: fuera del proyecto y de `.git`, nada que ignorar ni versionar. Perderla solo cuesta un `delta check` más. Un fallo al escribirla no cambia el resultado del check.

**Qué evita.** Cambiar un README o cualquier otro Markdown no pide un nuevo `delta check`. Un proyecto grande puede acotar el alcance de sus slices con `scope.<slice>` y así tampoco lo pide el código de otro slice; esa lista la mantiene una persona y, si se queda corta, `Stop` no ve un cambio que sí importaba.

**Límites.**
- *Shortcut: hay una pasada de git por slice, y `Tree` lee cada archivo cambiado del alcance; un binario grande sin seguir lo vuelve lento. Se afina calculando los cambios una vez y hasheando tamaño y fecha.*
- *Shortcut: "documentación" es el Markdown; se afina con patrones de ignorado en `deltas.config`.*
- Prueba que `delta check` corrió sobre este estado, no que el test se ejecutó: es el mismo límite que `Holds`.
- Dos árboles de trabajo del mismo repo en rutas distintas llevan marcas separadas.

### 8.8 `deltas.config`

Un archivo opcional, `deltas/deltas.config` en la raíz del proyecto (y solo ahí: un `deltas/` de servicio no trae el suyo), con líneas `clave=valor`; las líneas en blanco y las que empiezan con `#` se ignoran, y gana el último valor de una clave. Sin él todo funciona con los valores por defecto: `deltas.config` solo existe para lo que ata a Delta a un lenguaje, a una herramienta o a un tamaño de proyecto. **No** configura `deltas/`, `specifies/` ni `.spec.md`: son el formato del producto. El que instala `delta init` escribe esos valores por defecto en vez de dejarlos comentados (salvo `scope.<slice>`, que no tiene un valor por defecto que escribir), separados por líneas en blanco y con el presupuesto agrupado por clase de carga; una prueba exige que coincidan con los del código.

| Clave | Qué hace | Por defecto |
|---|---|---|
| `test_files` | Patrones (como `path.Match`, sobre el nombre del archivo, separados por coma o espacio) de los archivos que son tests. Sustituyen a la regla propia, que es esta misma lista; un `.md` nunca es un test | `*_test.go`, `test_*.py`, `*_test.py`, `*.test.*` y `*.spec.*` |
| `command_timeout` | Tiempo máximo de cada comando de un claim, como `90s` o `10m` | `10m` |
| `required_tools` | Programas que deben estar en el `PATH`; `delta status` y `SessionStart` avisan de los que falten | ninguno |
| `scope.<slice>` | Rutas de las que depende el slice; reduce su alcance en `Stop` (sección 8.7) | todo el código |
| `always_tokens`, `claude_lines`, `rule_always_lines`, `skill_desc_chars`, `agent_desc_chars` | Los techos de `delta budget` (sección 8.6) | los de la sección 8.6 |

**Qué pasa con una línea inválida.** `conffile.Parse` nunca falla: deja fuera la línea y la cuenta en `Conf.Problems` (`deltas/deltas.config:3: unknown key "colour"`), y `delta status` y `SessionStart` la muestran como aviso. Así un error de tipeo no rompe un hook.

**Cómo se usa.** La raíz de composición (`scripts/delta/run.go`) lo lee una vez y entrega a cada pieza lo suyo: los patrones a `vaults`, el tiempo a `runner`, los techos a `measure`, los alcances a `status` y `verify`, y las herramientas y los problemas a `setup`. Si no hay `deltas.config` pero sí el `.claude/budget.conf` de la generación anterior, se toman de él los techos. *Shortcut: esa lectura se quita cuando `delta init` instale `deltas.config` en lugar de `budget.conf`.*

**Lo que no hay, a propósito.** Sin plantillas de comando con `{id}` o `{paths}`: el filtro de un test cambia entre runners (`-run` en Go, `-t` en jest) y la regla de que el comando nombre sus rutas se rompería. Se añade una clave solo cuando una necesidad real la pida. Los comandos corren desde la carpeta de su scope (sección 16).

---

## 9. Librerías

| Necesidad | Qué usar |
|---|---|
| Comandos, ayuda, subcomandos | **`spf13/cobra`** (la única dependencia de terceros decidida) |
| Ejecutar comandos | `os/exec` y `context` |
| Partir un comando en argumentos | Función propia de unas 30 líneas, o `google/shlex`. Sin expansión de variables; ojo con las comillas en rutas con espacios de Windows |
| Leer specs | `bufio` y `regexp`, parser propio |
| Frontmatter | Parser propio de `clave: valor` plano |
| Plantillas y harness | `embed` y `text/template` |
| Tablas en pantalla | `text/tabwriter` |
| Salida para máquinas | `encoding/json` |
| Git | El binario `git` por `os/exec` |
| Pruebas | `testing`, golden files en `testdata/`; opcionalmente `google/go-cmp` |

**Qué se descarta y por qué:**

- `go-git`: enorme para lo que se necesita (un `diff --name-only` y leer una configuración).
- Un parser de Markdown (goldmark u otro): el formato es fijo y se necesitan los números de línea.
- Una librería de YAML: el frontmatter son tres claves planas. *Shortcut: solo claves planas. Si el frontmatter crece, se cambia por una librería YAML mantenida (hay que revisar cuál lo está en ese momento).*
- `testify`: la librería estándar y `go-cmp` alcanzan.
- `doublestar`: los tests se reconocen por el nombre del archivo (`*_test.go`, `*.test.*`, `*.spec.*`), sin patrones `**`.

---

## 10. Reglas para escribir el código

El proyecto debe ser limpio y robusto, y se aplica a sí mismo lo que Delta predica.

- **Parsea, no valides.** Los constructores devuelven un tipo válido o un error (`claim.ParseID`). Un `claim.ID` ya creado no puede ser inválido, así que no se vuelve a comprobar.
- **Errores con contexto.** Se envuelven con `%w`. Los errores de parseo llevan archivo y línea.
- **Sin globales ni `init()`.** Todo por constructor. `context.Context` es el primer parámetro de todo lo que hace I/O.
- **Interfaces chicas, declaradas por quien las usa.** Una o dos funciones.
- **El dominio es puro.** Sin reloj, sin disco, sin variables de entorno. Lo que necesite del mundo, lo recibe como argumento.
- **Salida determinista.** Nada de recorrer mapas en orden aleatorio; todo se ordena antes de imprimir, para que los golden files sean estables.
- **Comentarios solo para el porqué**, en una línea. No se comenta lo que el nombre o la firma ya dicen, tampoco en los tipos y funciones exportados ni en cada constante de un grupo. Cuando un atajo tiene un límite conocido: `shortcut: <el límite>, <cuándo mejorarlo>`.
- **Tamaño.** Cada paquete hace una cosa; si un archivo necesita dos secciones de comentarios para explicarse, son dos paquetes.
- **Legibilidad antes que brevedad.** Una función cabe en una pantalla (unas 30 líneas) y hace una cosa; un archivo agrupa una responsabilidad (`specfile` se lee en `header.go`, `claims.go`, `naming.go` y `parsing.go`); en el adaptador, un comando por archivo. Preferir el flujo plano con retornos tempranos a los `if` anidados, y un `switch` ordenado a una cadena de `else`.
- **Nombres que no tartamudean ni chocan.** El tipo principal de un paquete no repite su nombre (`verify.Checker`, no `verify.Verify`) y no comparte nombre con un tipo de otro paquete (`status.Reporter`, no `status.Status`, que es `lifecycle.Status`). `New` solo construye: lo que crea un spec se llama `Create`.
- **Nada que no se use.** Una función exportada que solo llaman los tests se borra, y los tests prueban lo que el código usa de verdad.
- **Herramientas.** `go vet`, `golangci-lint` y `gofmt` en CI; tests con `-race`.

---

## 11. Pruebas

| Qué | Cómo | Capa |
|---|---|---|
| Reglas del dominio | Tests de tabla; la mayoría de los casos se resuelven aquí, sin disco | `domain` |
| Parser de specs | Tests de tabla, golden files y **fuzzing** | `specfile` |
| Casos de uso | Con puertos falsos escritos a mano (sin librería de mocks) | `usecase` |
| Gateways | Contra un directorio temporal real y un repo git temporal | `gateway` |
| Salida de comandos | Golden files del texto y del JSON | `adapter` |
| Arquitectura | El test de imports de la sección 4 | todo `source` |
| Compatibilidad de `delta check` | Proyectos mínimos reales en Go y TypeScript con un test que pasa, uno que falla, un comando que sale con 0 sin test que cite el ID y un `skip` (`scripts/delta/testdata/compat/{go,ts}`) | extremo a extremo |

La CI corre en Linux, macOS y **Windows** desde el primer día: rutas con espacios, finales de línea y la forma de partir comandos son donde más fácil se rompe.

Todas estas pruebas son gratis y no usan ninguna API: es la "fiabilidad del sistema" que `direction.md` promete.

---

## 12. `source/harness/` y su relación con `.claude/`

`source/harness/` es la **fuente** de lo que Delta instala en un proyecto. `delta init` lo escribe en el proyecto destino:

| En `source/harness/` | Se instala como |
|---|---|
| `skills/<skill>/**` | `.claude/skills/<skill>/**`; quill, forge y seal. Los `.sh`, si hubiera, se instalan ejecutables |
| `rules/**/*.md` | `.claude/rules/<nombre>.md` (las subcarpetas `core/` y `base/` de la Fase 4 se aplanan) |
| `core/githooks/commit-msg` | `.githooks/commit-msg` |
| `core/configs/settings.json` | fragmento fusionado en `.claude/settings.json` |
| `core/configs/deltas.config` | `deltas/deltas.config` (sección 8.8) |
| `core/templates/CLAUDE.md` | el `CLAUDE.md` mínimo, con el nombre del proyecto |
| `core/templates/spec.md` | no se instala: lo usa `delta spec new` |

**El detalle del embed.** `go:embed` solo ve la carpeta de su paquete y sus subcarpetas, así que `source/gateway/bundle/` no puede embeber la carpeta hermana `source/harness/`. Por eso `source/harness/harness.go` declara el `embed.FS` y `gateway/bundle` lo importa. El test de capas trata `harness` como datos: no importa nada del módulo y solo `gateway` lo lee. Además, `embed` ignora por defecto los archivos que empiezan con `.` o `_`, por eso la carpeta se llama `core/githooks/` y se instala con el nombre con punto. Lo mismo vale para cualquier plantilla de un archivo con punto.

**Cómo se instala** (`usecase/setup`, con `gateway/bundle` como puerto):

- Un archivo que el proyecto ya tiene y que difiere del del harness **no se toca**: `delta init` lo nombra ("left as they are") y sigue. Uno idéntico se cuenta como ya instalado. Por eso correr `init` otra vez no cambia nada.
- `.claude/settings.json` es la excepción: se **fusiona**. Se añaden los grupos de hooks que no corran ya los mismos comandos con el mismo `matcher` (un `timeout` distinto es decisión del proyecto), las reglas de permisos que falten y las claves de primer nivel que no existan. No se quita ni se cambia nada; al reescribir el archivo se ordenan sus claves. Si el archivo no es un JSON con la forma esperada, se deja como está y se dice.
- Antes de escribir nada se fija `core.hooksPath`, de modo que fuera de un repositorio git `init` falla sin dejar archivos.

**`deltas/deltas.manifest`.** El nombre queda reservado para el registro de lo que `delta init` instaló (ruta y hash de cada archivo del harness), que hará falta para un futuro `delta update`: sin él `init` no distingue un archivo viejo de uno que el proyecto editó (hoy deja ambos como están). Todavía no se escribe ni se lee; `knowledge.Covers` ya lo trata como parte de la KB.

**Este repositorio no usa lo que instala** (decisión 10): `source/harness/` es la fuente y `delta init` lo escribe en *otros* proyectos. Aquí no hay `.claude/`, `deltas/`, `deltas.config` ni hook de git, y los tests lo prueban en repositorios temporales: instalan el harness, comprueban que `delta budget` pasa sobre él y que `delta status` funciona. Para cambiar una regla o un skill se cambia el harness.

---

## 13. Distribución

*Propuesta, sin decidir los detalles.*

- `go install github.com/ronz204/deltas-ai/scripts/delta@latest` produce el binario `delta`.
- Binarios precompilados para Linux, macOS y Windows en las releases de GitHub, para quien no tenga Go.
- La versión se inyecta en la compilación con `-ldflags "-X main.version=..."` y se muestra con `delta version`, que existe y hoy imprime `dev`. Su primera línea, `delta (deltas-ai) <versión>`, es la que el hook `commit-msg` usa para distinguir este programa de git-delta, así que no cambia.
- El CLI y el formato de spec se versionan por separado: el CLI con su versión, el spec con `format: N`.

---

## 14. Orden de implementación (Fase 1)

Cada paso termina con sus tests antes de empezar el siguiente:

1. Esqueleto: `go.mod`, las tres carpetas, `main.go` mínimo y el **test de dependencias**.
2. `domain/assertion`, `domain/specfile` y `domain/lifecycle`, con el fuzzing del parser.
3. `domain/linkcheck`, `domain/judgment` y los casos de uso `verify` para `delta spec lint` y `delta lint`.
4. `gateway/runner`, `gateway/vaults` y `delta check`.
5. `domain/drift` no entra aquí: es de la Fase 5.
6. `gateway/gitcli`, `domain/commits`, `domain/budget` y sus comandos.
7. La suite de compatibilidad con proyectos reales de Go y TypeScript.

`adapter/hook`, `usecase/guard` y `setup` pertenecen a la Fase 2; `author` con `approve` y las transiciones, a la Fase 3.

---

## 15. Riesgos y límites

| Riesgo o límite | Qué hacer |
|---|---|
| El parser de specs es el cuello de botella de todo | Se escribe primero, con fuzzing y errores con archivo y línea |
| Windows: rutas con espacios, CRLF, partir comandos | CI en Windows desde el inicio; `.gitattributes` con LF |
| El formato del JSON de los hooks puede cambiar | El adaptador `hook` no tiene reglas; si cambia la plataforma, se reescribe solo ese paquete |
| La estructura de cuatro capas es más pesada que un `main` y unos archivos | Es la que hace posible el test de dependencias; si en la práctica estorba, `usecase` y `domain` se pueden fundir sin tocar `gateway` ni `adapter` |
| Un `Holds` no prueba que el test corrió | Límite documentado; mejora futura: archivo de evidencia (`direction.md`, §10) |
| `stale` por archivo da falsos positivos | Medir el ruido en Kelsen antes de afinar (Fase 6) |
| El CLI crece sin control | Cada comando y cada paquete pasa la prueba de entrada (pilar 9) |

---

## 16. Pendientes menores

**Abiertos**

*Ninguno por ahora.*

**Decididos, por implementar**

*Ninguno por ahora.*

**Resueltos**

- **Un solo `deltas/`.** Antes cada carpeta `deltas/` del árbol era un scope (slices `servicio/slice`, comandos desde la carpeta del scope, `linkcheck.Under`, `vaults.Resolve`). Se quitó para que la KB no se disperse: hay un `deltas/` en la raíz, un slice se llama con un nombre sin carpeta (`invoicing`) y el comando de un spec corre desde la raíz del repo. Un monorepo lo expresa en el comando (`go test ./services/auth`). Un `deltas/` en otra carpeta no se lee, y un nombre con barra no es un slice. *Lo que sigue en pie: `scope.<slice>` de `deltas.config` acota el código que mira un slice para `Stop`, que no tiene que ver con carpetas.*
- **Quién dispara `approved → built` y `built → verified`.** `delta built` (Forge) y `delta check`; tabla en la sección 7.2.
- **`spec write` y el estado.** Un spec nuevo solo nace `draft`, y solo los comandos de estado lo mueven hacia delante (sección 7.2).
- **Formato del JSON de los hooks.** Verificado en la documentación de Claude Code; el contrato y sus límites están en la sección 8.5.
- **Versión de Go y ruta del módulo.** Fijadas en `go.mod`: `github.com/ronz204/deltas-ai`, Go 1.27.
- **Configuración de `golangci-lint`.** `.golangci.yml` activa `errcheck`, `govet`, `ineffassign`, `staticcheck` y `unused`, y un trabajo `lint` de CI la corre. `errcheck` ignora las escrituras con `fmt.Fprint*`, `Flush` de `tabwriter` y `os.Remove`, que no tienen nada útil que manejar. Ya encontró una constante muerta en un test y un `Close` sin comprobar.
- **Códigos de diagnóstico.** No es una decisión sino una práctica: ya existen `SPEC-*`, `LINT-*`, `COMMIT-*`, `BUDGET-*` y `CHECK-RECORD`, y cada comando nuevo añade los suyos (sección 7.3).
- **`Kind` y `TagKind` en `assertion`.** `TagKind` pasó a llamarse `Method` y el campo `Tag.Kind` a `Tag.Method`, para no confundirlo con `ID.Kind()` (invariante o aceptación).
- **El nombre del binario.** Se conserva `delta`. `delta version` imprime `delta (deltas-ai) <versión>` en su primera línea, y el hook `commit-msg` la comprueba antes de ejecutar nada: si el `delta` del `PATH` es otro (git-delta), avisa y deja pasar el commit en vez de cancelarlo. `delta status` y `SessionStart` también avisan (`setup.Check`).
- **`Stop` y la evidencia.** Sección 8.7: huella del alcance de cada slice, marca en la caché del usuario y un solo bloqueo.
- **`PostToolUse`.** Corre `delta lint` cuando Claude escribe un archivo de test y le cuenta los problemas de ese archivo (sección 8.5). Los linters de lenguaje los maneja cada proyecto.
- **`deltas.config`.** Sección 8.8: patrones de archivos de test, tiempo límite, herramientas requeridas, alcance por slice y los techos del presupuesto. `deltas/`, `specifies/` y `.spec.md` siguen fijos, y `.claude/` lo conocen solo `budget` y `hook`.
- **Qué mira `Stop`.** Por defecto, todo el código del proyecto; se acota por slice con `scope.<slice>` en `deltas.config`. Un comando que nombra rutas estrechas no decía qué código alcanza, y un recordatorio que a veces sobra es mejor que uno que a veces falla sin avisar.
