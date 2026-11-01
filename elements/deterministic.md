# Delta: la parte determinista

*Versión 1, 1 de noviembre de 2026. Diseño del CLI `delta` en Go: estructura, capas, paquetes, librerías y reglas de código.*

Este documento explica cómo se construye la parte determinista del harness: todo lo que un programa puede decidir sin un modelo. Complementa a `redirection.md` (qué es Delta y por qué), `background.md` (contexto) y `roadmap.md` (en qué orden). Si algo choca con `redirection.md`, manda `redirection.md`.

**Cómo leerlo:** lo marcado **Decidido** salió de la conversación con Roney y se da por cerrado. Lo marcado *propuesta* es mi criterio y se puede cambiar. Lo marcado *a verificar* es algo que no comprobé y hay que confirmar antes de implementar. Ningún código de este documento existe todavía: los fragmentos de Go son bocetos para fijar los nombres y las formas, no una implementación.

---

## 1. Qué es y qué no es

La idea central de Delta es **el modelo propone, el código decide** (`redirection.md`, §1). Esta parte es el "código decide":

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
| 2 | Tres carpetas de código: `entries`, `internal`, `harness` | Decidido |
| 3 | `go.mod` en la raíz del repo | Decidido |
| 4 | Cuatro capas dentro de `internal`: `domain`, `usecase`, `gateway`, `adapter` | Decidido |
| 5 | Nombres de paquete de una palabra, de 4 a 6 letras | Decidido |
| 6 | CLI con `spf13/cobra` | Decidido |
| 7 | Se añade el comando `delta hook <evento>` | Decidido |
| 8 | Los comandos de verificación se ejecutan directo con `os/exec`, sin shell | Decidido (`redirection.md`, 6.4) |
| 9 | `delta check` no lee la salida de los runners: `Holds` exige exit 0 más un test que cite el ID | Decidido (`redirection.md`, 6.4 y 6.5) |
| 10 | `.claude/` de este repo no se toca hasta que `delta init` pueda regenerarlo; `harness/` nace como copia renombrada | Decidido |
| 11 | Frontmatter plano (`clave: valor`) con parser propio; parser de claims propio | Propuesta |
| 12 | Salvo cobra, solo la librería estándar (sección 9) | Propuesta |

---

## 3. Estructura del repositorio

```
deltas-ai/
├── go.mod
├── entries/
│   └── delta/main.go            # raíz de composición: crea las piezas y las conecta
├── internal/
│   ├── domain/                  # puro: sin I/O, solo librería estándar
│   │   ├── claim/               # Claim, ID, Tag, enunciado EARS
│   │   ├── spec/                # Spec, Frontmatter, versión de formato, parser y render
│   │   ├── slice/               # Slice, Status, transiciones válidas
│   │   ├── proof/               # Verdict (Holds, Violated, Unverified) y la regla de evidencia
│   │   ├── drift/               # regla de stale: archivos cambiados vs referenciados
│   │   ├── lint/                # reglas de coherencia entre specs y tests
│   │   ├── commit/              # validador de Conventional Commits
│   │   └── budget/              # conteo de caracteres y techo
│   ├── usecase/                 # orquestación; declara los puertos (interfaces) que necesita
│   │   ├── setup/               # delta init
│   │   ├── author/              # spec new, spec write, link, approve
│   │   ├── verify/              # lint, check
│   │   ├── status/              # status, stale
│   │   └── guard/               # decisiones de hooks: qué se permite, qué se bloquea
│   ├── gateway/                 # los únicos efectos: implementan los puertos
│   │   ├── files/               # lectura y escritura atómica de la KB
│   │   ├── runner/              # os/exec sin shell, con timeout
│   │   ├── gitcli/              # envoltorio fino del binario git
│   │   └── bundle/              # el harness embebido, listo para instalar
│   └── adapter/                 # traducen el mundo exterior a casos de uso
│       ├── cli/                 # comandos cobra y formato de salida
│       └── hook/                # JSON de Claude Code por stdin → exit code
└── harness/                     # lo que `delta init` instala (se embebe en el binario)
    ├── harness.go               # //go:embed
    ├── skills/                  # quill, forge, seal
    ├── rules/                   # core (R1, R2) y base (B1, B2)
    ├── githooks/                # commit-msg (se instala como .githooks/)
    ├── settings/                # fragmento de hooks para .claude/settings.json
    └── templates/               # spec.md, CLAUDE.md
```

**Por qué `entries` y no `cmd`:** es una preferencia de nombre. `go install ./entries/delta` sigue produciendo un binario llamado `delta`, porque el nombre sale de la carpeta que contiene el `main`.

**Por qué `internal`:** el compilador impide que otros módulos importen estos paquetes. La API pública de Delta es el CLI y el formato del spec, no los paquetes Go.

**Lo que no está en esta estructura:** `elements/` (documentos de trabajo), `deltas/` y `.claude/` siguen en la raíz como están. `.claude/` es la configuración con la que se trabaja hoy en este repo; se regenerará desde `harness/` cuando exista `delta init` (sección 11).

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

La raíz de composición, `entries/delta/main.go`, es el único lugar que conoce todas las capas. Crea los gateways, se los pasa a los casos de uso por constructor, y los casos de uso a los adaptadores. No hay variables globales ni `init()`.

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
| `claim` | El átomo del contrato | `ID`, `ParseID(string) (ID, error)`, `Tag` (`Test` con comando, `Manual`, `None`), `Claim` (ID, tag, texto, línea) |
| `spec` | El archivo completo | `Spec` (formato, slice, estado, claims), `Parse(name, src) (Spec, []Diagnostic)`, `SetStatus(src, status) ([]byte, error)`, `SetTag(src, id, tag) ([]byte, error)` |
| `slice` | El ciclo de vida | `Status` (`Draft`, `Approved`, `Built`, `Verified`, `Stale`), `Can(from, to) bool` |
| `proof` | El resultado de verificar | `Verdict` (`Holds`, `Violated`, `Unverified`), `Outcome` (claim, veredicto, motivo), `Judge(tag, exit, cited) Outcome` |
| `drift` | Cuándo un contrato está viejo | `Stale(referenced, changed []string) bool` y qué archivos coinciden |
| `lint` | Coherencia entre specs y tests | Reglas: ID sin test que lo cite, test que cita un ID inexistente, claim sin tag, `skip`/`only` en archivos que citan IDs |
| `commit` | Conventional Commits | `Validate(message) []Diagnostic`: tipos permitidos, longitud de 40 a 50, imperativo, minúscula, sin punto final |
| `budget` | Presupuesto de contexto | `Count(text) int` (caracteres entre 4), `Check(total, ceiling) Outcome` |

Observaciones:

- **`spec` edita por líneas.** Cambiar el estado o el tag de un claim modifica únicamente esa línea y deja el resto del archivo intacto. Es lo que permite que `delta link` y `delta approve` toquen un solo renglón.
- **`Judge` es la regla de evidencia en una función:** exit distinto de 0 es `Violated`; exit 0 con un test que cita el ID es `Holds`; exit 0 sin test que cite el ID es `Unverified` con el motivo escrito; `manual` y `none` son `Unverified`. Es la pieza más fácil de probar de todo el proyecto.
- **`lint` reconoce el ID con coincidencia tolerante** (`INV-01`, `INV_01`, `INV01`, con un límite al final para que `INV01` no coincida con `INV011`), porque en Go los nombres de función no admiten guiones.

### 5.2 `usecase`

| Paquete | Comandos | Qué orquesta |
|---|---|---|
| `setup` | `delta init` | Escribe el esqueleto, instala el harness desde `bundle`, fija `core.hooksPath` sin pisar uno ajeno. Es idempotente |
| `author` | `delta spec new`, `spec write`, `link`, `approve` | Valida con `domain` y escribe con `files`. Es el único camino de escritura de la KB |
| `verify` | `delta spec lint`, `lint`, `check` | Carga specs, busca tests que citen IDs, ejecuta comandos con `runner`, aplica `proof.Judge` |
| `status` | `delta status`, `stale` | Reúne estado, cobertura y deriva por slice; compara con `git` |
| `guard` | `delta hook`, `commit-msg`, `budget` | Responde "¿se permite esto?" para cada evento: editar la KB, commitear, terminar la sesión |

### 5.3 `gateway`

| Paquete | Efecto | Detalles |
|---|---|---|
| `files` | Lee y escribe la KB | Escritura atómica: archivo temporal en la misma carpeta y renombrado. Conserva LF. Localiza la raíz del repo |
| `runner` | Ejecuta comandos | `os/exec` con `context` y timeout. Parte el comando en argumentos sin expandir variables. *Shortcut: sin shell, no admite pipes ni redirecciones; si hacen falta, se escribe un programa* |
| `gitcli` | Llama a `git` | `diff --name-only`, leer y fijar `core.hooksPath`. Se usa el binario de git, no una librería |
| `bundle` | Entrega el harness | Envuelve el `embed.FS` de `harness/` y sabe a qué ruta del proyecto va cada archivo |

### 5.4 `adapter`

| Paquete | Entrada | Salida |
|---|---|---|
| `cli` | Argumentos y flags (cobra) | Texto en tabla con `text/tabwriter`, o JSON con `--json`; exit code |
| `hook` | JSON del evento de Claude Code por stdin | Exit 0 para permitir, exit **2** para bloquear (el 1 no bloquea), con el motivo en stderr |

`hook` no contiene reglas. Extrae del JSON lo que `guard` necesita (herramienta, ruta, comando), llama a `guard` y traduce la decisión a un exit code.

---

## 6. De los comandos a los paquetes

| Comando | Adaptador | Caso de uso | Dominio que decide |
|---|---|---|---|
| `delta init` | `cli` | `setup` | `slice`, `spec` (plantilla) |
| `delta spec new <slice>` | `cli` | `author` | `spec`, `slice` |
| `delta spec lint [slice]` | `cli` | `verify` | `spec`, `claim` |
| `delta spec write` | `cli` | `author` | `spec` |
| `delta link <claim-id> <comando>` | `cli` | `author` | `spec`, `claim` |
| `delta lint` | `cli` | `verify` | `lint`, `claim` |
| `delta check [slice]` | `cli` | `verify` | `proof`, `lint` |
| `delta status` | `cli` | `status` | `slice`, `proof`, `drift` |
| `delta stale` | `cli` | `status` | `drift` |
| `delta budget` | `cli` | `guard` | `budget` |
| `delta commit-msg <archivo>` | `cli` | `guard` | `commit` |
| `delta approve <slice>` | `cli` | `author` | `slice` |
| `delta context <slice>` | `cli` | `status` | `spec`, `claim` |
| `delta hook <evento>` | `hook` | `guard` | `slice`, `commit`, `budget`, `proof` |

`delta hook` se invoca desde `.claude/settings.json` con un evento (`session-start`, `pre-edit`, `pre-commit`, `post-edit`, `stop`). *Propuesta de nombres; se fijan al implementarlo.*

---

## 7. Contratos de datos

### 7.1 El spec

Definido en `redirection.md`, 6.4. Resumen de lo que el parser exige:

- Frontmatter con `format`, `slice` y `status`, claves planas.
- Cada claim es una línea `- <ID> [<tag>]` seguida de su texto indentado.
- IDs únicos dentro del spec. Tags: `test: <comando>`, `manual`, `none`.
- `format: 1` permite que una versión futura migre specs viejos.

### 7.2 Estados del slice

```
draft  ->  approved  ->  built  ->  verified
                           ^           |
                           +-- stale <-+
```

`slice.Can(from, to)` es la única fuente de qué transiciones son legales. Solo el CLI modifica el estado.

### 7.3 Diagnósticos y códigos de salida

- Un **diagnóstico** lleva archivo, línea, un código estable por tipo (por ejemplo `SPEC-DUP-ID`) y un mensaje. Hooks y CI pueden depender del código, no del texto.
- **Códigos de salida del CLI:** 0 todo bien, 1 se encontraron violaciones, 2 uso incorrecto.
- **En `delta hook`:** 0 permite y 2 bloquea, porque eso es lo que Claude Code espera.
- Con `--json` la salida es estable y ordenada, pensada para hooks y CI.

---

## 8. Las piezas críticas, paso a paso

### 8.1 El parser de specs

Es la pieza de la que depende todo lo demás, así que se escribe primero. Trabaja línea por línea con `bufio` y `regexp`; no usa un parser de Markdown general porque el formato es un subconjunto fijo y un parser general escondería los números de línea de los errores. Devuelve el `Spec` y la lista completa de diagnósticos, no solo el primero. Tiene fuzzing nativo de Go, porque es lo que más datos ajenos recibe.

### 8.2 `delta check <slice>`

1. `verify` carga el spec y lo parsea. Si hay diagnósticos de error, termina con ellos.
2. Para cada claim con tag `test`, busca en los archivos de tests si alguno cita el ID (`lint`).
3. `runner` ejecuta el comando desde la raíz del repo y devuelve el código de salida. La salida del comando se muestra a la persona pero **no se interpreta**.
4. `proof.Judge` combina tag, código de salida y "¿algún test cita el ID?" en un `Outcome`.
5. Los claims `manual` y `none` se reportan como `Unverified`.
6. Se imprime la tabla y se sale con 1 si hay algún `Violated`.

Qué garantiza y qué no: un `Holds` prueba que el comando salió con 0 y que existe un test que cita el ID. **No prueba que ese test se ejecutó** (un `skip` o un comando que apunta a otro archivo pasarían). Ese límite está documentado en `redirection.md`, §10, junto con la mejora futura (un archivo de evidencia que cada test escribe).

### 8.3 `delta stale`

Para cada spec, `drift` compara los archivos que el spec referencia con los que `gitcli` dice que cambiaron desde la última verificación. La granularidad es el archivo (decisión de `redirection.md`). Marca de más si el cambio no altera el comportamiento, y no marca si el comportamiento cambia sin tocar los archivos nombrados. Está documentado como heurística.

### 8.4 Escritura atómica de la KB

`author` valida con `domain`, y solo si no hay errores `files` escribe en un archivo temporal de la misma carpeta y lo renombra sobre el destino. Un spec nunca queda a medias.

### 8.5 `delta hook`

Cada evento sigue el mismo camino: `hook` lee el JSON, `guard` decide, `hook` traduce a exit code.

| Evento | Qué decide `guard` |
|---|---|
| `session-start` | Qué slices están `stale` o sin verificar; avisa si falta `core.hooksPath` o el binario |
| `pre-edit` | Bloquea la edición directa de la KB por el agente |
| `pre-commit` | Valida el mensaje con `commit.Validate`; bloquea con 2 si no cumple |
| `post-edit` | Dispara el linter del proyecto y `stale` |
| `stop` | Exige evidencia de que `delta check` corrió tras el último cambio |

*A verificar:* el formato exacto del JSON que Claude Code entrega a cada hook (nombres de campos, cómo llega el comando de Bash) y cómo declarar los eventos en `settings.json`. No lo leí de la documentación vigente; hay que confirmarlo antes de escribir `adapter/hook`. Es el punto más expuesto a cambios de plataforma, y por eso el adaptador se mantiene delgado.

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
- **Comentarios solo para el porqué**, en una línea. Cuando un atajo tiene un límite conocido: `shortcut: <el límite>, <cuándo mejorarlo>`.
- **Tamaño.** Cada paquete hace una cosa; si un archivo necesita dos secciones de comentarios para explicarse, son dos paquetes.
- **Herramientas.** `go vet`, `golangci-lint` y `gofmt` en CI; tests con `-race`.

---

## 11. Pruebas

| Qué | Cómo | Capa |
|---|---|---|
| Reglas del dominio | Tests de tabla; la mayoría de los casos se resuelven aquí, sin disco | `domain` |
| Parser de specs | Tests de tabla, golden files y **fuzzing** | `spec` |
| Casos de uso | Con puertos falsos escritos a mano (sin librería de mocks) | `usecase` |
| Gateways | Contra un directorio temporal real y un repo git temporal | `gateway` |
| Salida de comandos | Golden files del texto y del JSON | `adapter` |
| Arquitectura | El test de imports de la sección 4 | todo `internal` |
| Compatibilidad de `delta check` | Proyectos mínimos reales en Go y TypeScript con un test que pasa, uno que falla, un comando que sale con 0 sin test que cite el ID y un `skip` | extremo a extremo |

La CI corre en Linux, macOS y **Windows** desde el primer día: rutas con espacios, finales de línea y la forma de partir comandos son donde más fácil se rompe.

Todas estas pruebas son gratis y no usan ninguna API: es la "fiabilidad del sistema" que `redirection.md` promete.

---

## 12. `harness/` y su relación con `.claude/`

`harness/` es la **fuente** de lo que Delta instala en un proyecto. `delta init` lo escribe en el proyecto destino:

| En `harness/` | Se instala como |
|---|---|
| `skills/{quill,forge,seal}/` | `.claude/skills/...` |
| `rules/core/`, `rules/base/` | `.claude/rules/...` |
| `githooks/commit-msg` | `.githooks/commit-msg` |
| `settings/` | fragmento fusionado en `.claude/settings.json` |
| `templates/spec.md`, `templates/CLAUDE.md` | un spec nuevo y el `CLAUDE.md` mínimo |

**El detalle del embed.** `go:embed` no puede subir de carpeta, así que un paquete en `internal/gateway/bundle/` no puede embeber `harness/` desde la raíz. Por eso `harness/harness.go` declara el `embed.FS` y `gateway/bundle` lo importa. Además, `embed` ignora por defecto los archivos que empiezan con `.` o `_`, por eso la carpeta se llama `githooks/` y se instala con el nombre con punto. Lo mismo vale para cualquier plantilla de un archivo con punto.

**Convivencia con el `.claude/` actual** (decisión 10):

1. `.claude/` de este repo no se toca. Contiene las reglas que rigen el trabajo aquí (commits, lean-change).
2. En la Fase 0, `harness/` nace como copia ya renombrada (Quill, Forge, Seal).
3. Cuando `delta init` exista (Fase 2), se regenera `.claude/` desde `harness/` y se elimina la duplicación. De paso, el repo prueba su propio instalador.

---

## 13. Distribución

*Propuesta, sin decidir los detalles.*

- `go install github.com/ronz204/deltas-ai/entries/delta@latest` produce el binario `delta`.
- Binarios precompilados para Linux, macOS y Windows en las releases de GitHub, para quien no tenga Go.
- La versión se inyecta en la compilación con `-ldflags` y se muestra con `delta version`.
- El CLI y el formato de spec se versionan por separado: el CLI con su versión, el spec con `format: N`.

---

## 14. Orden de implementación (Fase 1)

Cada paso termina con sus tests antes de empezar el siguiente:

1. Esqueleto: `go.mod`, las tres carpetas, `main.go` mínimo y el **test de dependencias**.
2. `domain/claim`, `domain/spec` y `domain/slice`, con el fuzzing del parser.
3. `domain/lint`, `domain/proof` y los casos de uso `verify` para `delta spec lint` y `delta lint`.
4. `gateway/runner`, `gateway/files` y `delta check`.
5. `domain/drift` no entra aquí: es de la Fase 5.
6. `gateway/gitcli`, `domain/commit`, `domain/budget` y sus comandos.
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
| Un `Holds` no prueba que el test corrió | Límite documentado; mejora futura: archivo de evidencia (`redirection.md`, §10) |
| `stale` por archivo da falsos positivos | Medir el ruido en Kelsen antes de afinar (Fase 6) |
| El CLI crece sin control | Cada comando y cada paquete pasa la prueba de entrada (pilar 9) |

---

## 16. Pendientes menores

1. **Quién dispara `approved → built` y `built → verified`.** `redirection.md`, 6.3, dice que solo el CLI cambia el estado, pero no qué comando lo hace. *Propuesta:* `delta check` mueve a `verified` cuando no queda ningún claim `test` en `Violated` ni `Unverified`, con los `manual` y `none` visibles como deuda; y Forge, al cerrar su trabajo, marca `built` con un comando. Se decide al implementar la Fase 3.
2. **Archivo de configuración.** `delta init` menciona "config", pero no está definido qué se configura. *Propuesta:* sin archivo de configuración en la primera versión, solo convenciones y valores por defecto (por ejemplo, los patrones de archivos de tests). Se añade si aparece una necesidad real.
3. **Formato del JSON de los hooks** y nombres definitivos de los eventos de `delta hook` (*a verificar*, sección 8.5).
4. **Versión de Go y ruta del módulo** (`github.com/ronz204/deltas-ai` es la ruta natural del repositorio actual; la versión de Go se fija al crear `go.mod`).
5. **Configuración de `golangci-lint`**: qué linters se activan.
6. **Códigos de diagnóstico**: la lista completa se define con el parser y las reglas de `lint`.
