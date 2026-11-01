# Delta: redirección

*Versión 3, 31 de octubre de 2026. Fija los nombres de los skills (Quill, Forge, Seal), los hooks en capas con `core.hooksPath` y un core de dos reglas con el resto en el preset `base`.*

*Nombres: en los datos del piloto y en "Dónde estábamos" se conservan los nombres viejos (Specifier, Executor, Sentinel). Del resto del documento en adelante, Specifier = **Quill**, Executor = **Forge**, Sentinel = **Seal**.*

Documento para decidir hacia dónde va Delta. Está escrito para leerse de punta a punta: primero el porqué, después el qué, y al final el cómo y los pendientes. El contexto completo (historia, investigación, conceptos) está en `backstory.md`.

**Cómo leerlo:** lo que viene del repo, de tu piloto y de los archivos que me pasaste es dato. Lo que viene de mi criterio está marcado como *propuesta*. Lo que no pude comprobar está marcado como *hipótesis*. Importante: no pude abrir la carpeta `.claude/` (el sitio bloquea el acceso automatizado), así que de los skills y scripts no leí nada directo. Trabajé con el README, tu `summary.md` y tu `code-lean-change.md`.

---

## 1. En una página

**Dónde estábamos.** Delta nació como una metodología de documentación viva para agentes, con cinco skills (Surveyor, Specifier, Archivist, Sentinel, Executor), un catálogo de reglas y un presupuesto de contexto. Se intentó justificar con ahorro de tokens, el piloto salió inconcluso (+29.4% de costo en tareas cortas, sin señal a favor) y quedó sin una pregunta clara que responder.

**A dónde vamos.** Delta deja de prometer ahorro y pasa a prometer **fiabilidad verificable**: lo que dice el contrato se puede comprobar, el código que se genera se rige por el contrato, y cuando ambos se separan, se nota solo.

**Cómo.** Con una idea central: **el modelo propone, el código decide.** Todo lo que se pueda expresar como regla sale de los prompts y pasa a un CLI en Go (`delta`) más hooks de Claude Code. El LLM queda para lo que de verdad es creativo o ambiguo: entrevistar, redactar claims, escribir el plan, escribir el código.

**Y la generación de código tiene reglas propias.** Un núcleo de solo dos reglas (fidelidad al contrato y honestidad), las únicas que se detectan con código y que dependen del contrato. Alcance y fiabilidad (la escalera y "lo que nunca se recorta") viven en el preset `base`, activado por defecto pero fuera del pitch: escribir mejor código es un **beneficio secundario** de Delta, no su promesa. Ponytail optimiza *cuánto código* se escribe; Delta optimiza *cuán fiel es el código al contrato*.

**Qué queda.** Tres roles con identidad propia, entrelazados con el ciclo del slice: **Quill** redacta el contrato (`draft`), **Forge** lo construye (`built`) y **Seal** lo certifica (`verified`). Además, los contratos vivos por slice, los tags de verificación y la regla de "sin bitácora".

**Qué se va.** Archivist y Surveyor como skills de LLM, el catálogo de reglas dentro del núcleo, `code-lean-change` como regla única (se disuelve en las reglas nuevas), la plantilla `ROADMAP.md`, la mención a Cursor y la promesa de ahorro de tokens.

**Qué se suma.** El CLI `delta`, hooks, IDs estables por claim, estados del slice, detección automática de deriva, dos reglas core propias, el preset `base` y presets con linters.

**Cómo sabemos que funcionó.** Sin benchmark obligatorio: usándolo en un repo real tuyo (Kelsen) durante varias features, y con tests normales sobre el CLI. Opcionalmente, medir fidelidad con `delta check`, sin juez LLM.

---

## 2. Qué aprendimos (la base de esta decisión)

### De tu piloto

- Costo: +25% a +41% en tareas cortas (+29.4% agregado). Costo fijo de ~4.2k tokens por sesión.
- Corrección idéntica en ambos brazos (23/23): la tarea ya saturaba, no había margen para mejorar.
- El spec no redujo la exploración, pero el fixture era de ~150 líneas: no había nada que explorar.
- Executor no se disparó en headless (0 de 6).
- La regla de commits se cumplió 1 de 3 veces.
- Tras construir un slice: 0 claims con tag `test`, 14 con `none`.
- Presupuesto: 3 473 de 3 500 tokens. Cualquier regla nueva lo revienta.

**Lectura honesta:** el piloto no prueba que Delta sea malo. Lo midió en el peor escenario posible y con un objetivo difuso. Lo que sí muestra son tres fallas de mecánica (reglas que no se cumplen, un skill que no se dispara, un lazo de verificación que nadie cierra), y las tres se arreglan con código, no con más prompts.

### De la investigación

- **Spec Kit y OpenSpec** imponen un proceso. Los críticos señalan mayor consumo de tokens, y no encontré evaluaciones rigurosas que demuestren mejor código. Su valor declarado es predictibilidad y trazabilidad, no ahorro.
- **Ponytail** es un conjunto de reglas de comportamiento con benchmark público (39 tareas, 5 corridas cada una, checks ocultos de corrección y seguridad, baseline sin skill, rivales incluidos). Su ahorro es sobre todo en tokens de *salida* (menos código escrito). Dos lecciones que nos sirven: cada regla tuvo que ganarle a la versión sin ella, y documentaron un error propio (un baseline contaminado por un hook).
- **Hooks vs CLAUDE.md:** CLAUDE.md, skills y rules son instrucciones que el modelo probablemente sigue; los hooks se ejecutan siempre. Un hook `PreToolUse` puede bloquear una acción (exit 2). Ojo: exit 1 *no* bloquea, y es el error clásico. Circula la cifra "70-90% de cumplimiento en prompts vs 100% en hooks", pero sale de un blog sin fuente: tomala como intuición, no como dato.
- **Rules con `paths:`** en el frontmatter se cargan solo cuando se tocan archivos que coinciden; las que no lo llevan cargan siempre.

**Conclusión:** el hueco que ninguna de esas herramientas llena es *mantener el conocimiento verdadero a lo largo del tiempo*. Spec Kit y OpenSpec guían el cambio; ponytail guía el estilo. Nadie detecta que el contrato quedó viejo. Ahí es donde Delta puede ser distinto.

---

## 3. La nueva dirección

### La tesis

> Delta mantiene un contrato verificable entre lo que el proyecto dice que hace y lo que el código realmente hace, guía al agente para que el código que genera sea fiel a ese contrato, y avisa cuando se separan.

Fíjate en lo que **no** dice: no dice "ahorra tokens", no dice "escribe mejor código que nadie". Dice que hay un contrato, que se puede comprobar, que el código se genera contra él y que la deriva no pasa en silencio.

### Pilares

| Pilar | Qué significa en concreto | Cómo se logra |
|---|---|---|
| **1. Fiabilidad** | Lo que se afirma se puede comprobar; lo que no, se declara como no verificado | Claims con ID y tag; `delta check` ejecuta los tests/checks enlazados; `Unverified` nunca finge ser `Holds` |
| **2. El código decide, el modelo propone** | Reglas, formato, estados y gates los aplica un programa, no un prompt | CLI en Go + hooks; el LLM no escribe la KB directo |
| **3. Código fiel al contrato** | Lo que el agente genera se rige por el spec: el contrato es el alcance, los tests citan claims, la ambigüedad vuelve al spec. Es parte de la fiabilidad, no una promesa de "mejor código" | Dos reglas core (R1 fidelidad, R2 honestidad, sección 7) apoyadas por `delta lint` y `delta check`. La calidad general de código queda en el preset `base` |
| **4. La deriva es ruidosa** | Un contrato viejo es peor que ninguno, porque el agente cree lo que lee | Hook que marca slices como `stale` cuando cambia el código que referencian |
| **5. Una sola fuente de verdad** | La KB no repite lo que el código ya dice | Solo se documenta lo que el código no muestra: intención, invariantes, decisiones |
| **6. El humano decide en los puntos caros** | Aprobar el plan antes de construir; el resto fluye | Estado `approved` obligatorio antes de que Forge arranque |
| **7. Claridad** | Cualquier persona (y cualquier agente) entiende el estado de un slice de un vistazo | `delta status`, specs cortos con estructura fija |
| **8. Adopción simple** | Funciona sin configurar nada | `delta init` con defaults sensatos; un binario sin runtime |
| **9. Cada pieza se gana su lugar** | Si quitar una regla no empeora nada, sobra | Prueba de entrada para reglas: "qué fallo previene y cómo se detecta" (idea tomada de ponytail) |

### Orden de prioridad cuando chocan

**Fiabilidad > claridad > adopción > costo.**

Ejemplo: más validación hace el flujo más confiable pero más pesado de adoptar. Con este orden gana la validación, y se compensa con buenos defaults (`delta init` que funcione sin preguntas). Si algún día querés que adopción pase por encima, el diseño cambia (menos gates, más avisos). Es una decisión tuya y conviene dejarla escrita.

El costo en tokens no es una meta, pero sí un límite: se carga lo que la tarea necesita, y el costo se ve.

### Qué se promete y qué no

Hay dos clases de fiabilidad y no hay que mezclarlas:

- **Fiabilidad del sistema** (el CLI valida, los estados se respetan, la deriva se detecta): se prueba con tests normales, gratis y sin API. **Esto sí se promete.**
- **Fiabilidad del resultado** ("el agente escribe mejor código con Delta", "mejor flujo y mantenibilidad"): es una intuición razonable, pero no está medida y el piloto no mostró señal. **Esto no se promete todavía**, aunque en la sección 13 hay una forma barata y determinista de medir una parte (fidelidad al contrato).

---

## 4. Qué se queda

| Concepto | Por qué se queda |
|---|---|
| **Contratos vivos, sin bitácora** | Es la idea fundacional: lo reemplazado desaparece, git guarda el pasado |
| **Slices** (un contrato por capacidad, dimensionado como bounded context) | Evita specs por capa o por archivo |
| **Claims en EARS + aceptación Given/When/Then** | Estructura fija = se puede validar con un parser |
| **Tags de verificación** (`test`, `manual`, `none`; desaparece `check`) | Es la base de la fiabilidad; ahora con IDs y con ejecución real |
| **Quill** (antes Specifier) | La entrevista es trabajo genuinamente de LLM |
| **Seal** (antes Sentinel) | Se queda como rol, pero se parte en dos (ver 6.5) |
| **Forge** (antes Executor) | Plan aprobado, tests primero, código después; nunca edita specs |
| **CLAUDE.md como router** | Sigue siendo la puerta de entrada, pero mínimo |
| **Docs curados, solo lo que el código no muestra** | Pilar 5 |
| **Un solo escritor de la KB** | Se mantiene, pero como comando (`delta spec write`) y no como skill |
| **Conventional Commits** | Se mantiene, pero aplicado por código en capas: hook de Claude Code para el agente, git hook para personas y CI (ver 6.2) |
| **Rules con alcance por ruta** | Se mantiene y se vuelve la norma |
| **Presupuesto de contexto** | Se mantiene, pero como gate y no como consejo |
| **Plan aprobado por la persona** | Punto de control humano |
| **Ideas rescatables de `code-lean-change`** | "El alcance lo marcan los claims", "lo que va más allá se nombra y se omite", "nunca recortar validación/errores/seguridad" y "cerrar diciendo qué quedó sin verificar". Se reubican en R1, R2 y el preset `base` (7.6) |

## 5. Qué se va (y por qué)

| Qué se va | Por qué |
|---|---|
| **Archivist como skill de LLM** | Era un LLM haciendo lo que un validador hace mejor. Su función ("único escritor, con la forma correcta") pasa al CLI, sin costo de tokens y sin variabilidad |
| **Surveyor como skill de LLM** | La parte determinista (crear la estructura) pasa a `delta init`. La parte creativa (redactar docs de orientación) nace a demanda, cuando Quill la necesita (decisión 2) |
| **El catálogo de reglas opcionales dentro del núcleo** | Pasa a "presets" opt-in fuera del camino principal. Cada regla se propone mostrando su costo en tokens |
| **La regla `commit-messages` como texto** (~1 079 t) | Es lo más caro del presupuesto y se cumplía 1 de 3 veces. Un git hook la cumple siempre y libera ese presupuesto |
| **`code-lean-change` como regla única** (~922 t) | Se disuelve en R1, R2 y el preset `base` (alcance, fiabilidad) y se reescribe con peldaños propios. Detalle en 7.6 |
| **Reglas chequeables como prosa** (por ejemplo "no uses `any`") | Pasan a linters. La rule en `.md` se reduce a una línea o desaparece |
| **`ROADMAP.md` plantilla** | Ningún skill lo lee; es peso sin función |
| **Cursor** | Ya lo habías eliminado a propósito, pero el About del repo todavía lo menciona. Hay que limpiar el README |
| **La promesa de ahorro de tokens como tesis** | No está medida y el piloto la contradice en repos chicos |
| **"Cinco roles" como mensaje principal** | Se reemplaza por tres roles más un CLI |

## 6. Qué se agrega

### 6.1 El CLI `delta` (Go)

Un binario único, sin runtime. Propuesta de comandos:

| Comando | Qué hace | Reemplaza |
|---|---|---|
| `delta init` | Crea la estructura base (`CLAUDE.md` router mínimo, `deltas/`, hooks de Claude Code, `.githooks/`, config) y fija `core.hooksPath` en el clon. Es idempotente: se corre de nuevo en cada clon | Parte determinista de Surveyor |
| `delta spec new <slice>` | Crea un spec desde la plantilla con frontmatter y estado `draft` | Parte de Archivist |
| `delta spec lint [slice]` | Valida estructura, notación EARS, IDs únicos y tags válidos | Revisión manual |
| `delta spec write` | Único punto de escritura de la KB; valida antes de escribir | Archivist |
| `delta link <slice> <claim-id> <comando>` | Enlaza un claim con el comando de su test (solo toca esa línea). Lleva el slice porque los IDs solo son únicos dentro de un spec | El paso que nadie podía hacer |
| `delta lint` | Cruza specs y tests: claims `test` cuyo ID no aparece en ningún archivo de tests, tests con ID de claim inexistente, claims sin tag; avisa de `skip`/`only` en archivos que citan IDs | Nada (es nuevo; respalda la regla de fidelidad) |
| `delta check [slice]` | Ejecuta el comando de cada claim `test` y reporta `Holds` / `Violated`; lo no ejecutable, o sin test que cite el ID, queda `Unverified` (ver 6.5) | La parte determinista de Sentinel (hoy Seal) |
| `delta status` | Tabla de slices: estado, cobertura, si está `stale` | `links.sh` ampliado |
| `delta stale` | Compara archivos modificados contra los que cada spec referencia y marca slices desactualizados | Nada (es nuevo) |
| `delta budget` | Mide lo siempre cargado y falla si pasa el techo | `budget.sh` como gate |
| `delta commit-msg <archivo>` | Valida un mensaje contra Conventional Commits (tipos permitidos, longitud, minúscula, sin punto final; el imperativo no se puede comprobar). Lo llaman el hook del agente, el git hook y CI | La regla `commit-messages` como texto |
| `delta hook` | Adaptador de los hooks de Claude Code: lee el JSON del evento por stdin (el evento va en el propio JSON, sin argumento), consulta la regla y sale con 0 o 2. No lleva lógica propia | Scripts sueltos por hook |
| `delta context <slice>` | Imprime solo lo necesario para trabajar ese slice | Carga bajo demanda real |
| `delta approve <slice>` | Pasa el spec a `approved` (lo hace la persona) | Punto de control humano formal |

Notas de diseño:
- Para el presupuesto alcanza con contar por caracteres (~4 caracteres por token). Un tokenizer exacto no vale la complejidad.
- El CLI no depende de Claude Code. Los hooks son un adaptador fino encima. Eso deja la puerta abierta a otros agentes sin que sea una meta. La estructura del proyecto Go, las capas y las librerías están en `blueprint.md`.

### 6.2 Hooks de Claude Code

| Evento | Qué hace |
|---|---|
| `SessionStart` | Inyecta el router mínimo y el resultado de `delta status` (qué slices están `stale` o sin verificar). Avisa si `core.hooksPath` no apunta a `.githooks` o si `delta` no está en el PATH |
| `PreToolUse` (Write/Edit sobre la KB) | Bloquea la edición directa de specs y docs por parte del agente (exit **2**, no 1). La persona sí puede editar a mano |
| `PreToolUse` (Bash con `git commit`) | Valida el mensaje con `delta commit-msg` y bloquea con exit **2** si no cumple. Es la capa que protege los commits del agente |
| `PostToolUse` (Write/Edit sobre código) | Corre el linter del proyecto y devuelve los errores al agente; además llama `delta stale` |
| `Stop` | Exige evidencia de que `delta check` corrió después del último cambio (regla de honestidad) y avisa si quedan slices `stale` o claims sin verificar |

Estos hooks viven en `.claude/settings.json`, que se versiona: **quien clone el repo los tiene sin configurar nada.**

#### Hooks de git y `core.hooksPath`

Git no versiona `.git/hooks/`, así que un hook ahí es personal de cada clon. Para compartirlos hay que versionarlos en una carpeta del repo y decirle a git que la use. Es una configuración local por clon, y no hay manera de evitarla. El diseño:

1. **Carpeta versionada `.githooks/`** con el hook `commit-msg`: un script POSIX `sh` mínimo que llama `delta commit-msg "$1"`. Git en Windows ejecuta los hooks con su propio `sh`, así que el mismo script sirve en las tres plataformas.
2. **`delta init` corre `git config core.hooksPath .githooks`** (local, idempotente). Es el único paso manual por clon, y si alguien lo olvida, no hay hook. Por eso existe el punto 3.
3. **La omisión no es silenciosa.** `delta status` y el hook `SessionStart` verifican que `core.hooksPath` valga `.githooks` y avisan si no.
4. **No se pisa lo ajeno.** Si el clon ya tiene un `core.hooksPath` distinto (Husky u otro), `delta init` no lo sobrescribe: avisa y explica cómo encadenar.
5. **Detalles de plataforma.** `.gitattributes` fija `.githooks/* text eol=lf` (un CRLF rompe el shebang en Unix). En Unix el bit ejecutable se registra con `git update-index --chmod=+x`.
6. **Si falta el binario.** Si `delta` no está en el PATH, el hook avisa y deja pasar el commit (*propuesta*: un hook que bloquea por una herramienta ausente estorba más de lo que protege). La ausencia la señala `SessionStart`.

Cuatro capas, de más a menos cercana al error:

| Capa | Protege a | Requiere setup por clon |
|---|---|---|
| Hook `PreToolUse` sobre `git commit` | El agente | No (va en `settings.json`) |
| Git `commit-msg` vía `core.hooksPath` | Las personas | Sí (`delta init`, una vez) |
| `delta commit-msg` en CI sobre el rango del PR | Todos, incluido `--no-verify` | No |
| Aviso en `SessionStart` y `delta status` | Detecta que falte la capa 2 | No |

`git commit --no-verify` salta el git hook; por eso el agente tiene su capa propia y CI es la red de seguridad.

### 6.3 Estados del slice

```
draft  ->  approved  ->  built  ->  verified
                           ^           |
                           +-- stale <-+   (el código o el spec cambian)
```

- Forge no arranca si el spec no está `approved` y limpio (`lint` sin errores).
- Un cambio al spec o al código referenciado después de `verified` devuelve el slice a `stale`.
- Los estados viven en el frontmatter del spec y solo los modifica el CLI.

Cada skill es dueño de un tramo del ciclo, y el nombre lo evoca:

| Skill | Imagen | Tramo | Qué hace |
|---|---|---|---|
| **Quill** | la pluma que redacta el contrato | hasta `draft` | Entrevista y redacta claims; la persona aprueba (`approved`) |
| **Forge** | la fragua donde se construye | `approved` → `built` | Plan, tests, código |
| **Seal** | el sello que certifica | `built` → `verified` | Verifica; si hay deriva, el sello se rompe (`stale`) y el ciclo vuelve a Quill |

Los tres nombres son sustantivos de 4-5 letras y los comandos quedan `/quill`, `/forge`, `/seal`. La descripción de cada skill (que es lo que dispara su carga) debe nombrar el verbo, no solo la metáfora.

### 6.4 IDs estables por claim

Sin ID no hay forma de enlazar tests, medir cobertura ni detectar que un claim desapareció (`Gone`). Es la pieza más barata y la que sostiene el resto, incluida la regla de fidelidad.

Formato propuesto (*borrador, a ajustar*):

```markdown
---
format: 1
slice: facturacion
status: approved
---

## Invariantes
- INV-01 [test: go test ./facturacion]
  Cuando se emite una factura, el sistema debe redondear cada línea a 2 decimales antes de sumar.
- INV-02 [none]
  El sistema no debe emitir dos facturas con el mismo consecutivo.

## Aceptación
- ACC-01 [test: go test ./e2e/consecutivo]
  Dado que existe la factura 1001, cuando se intenta emitir otra con 1001, entonces se rechaza.
```

**Tags de verificación: `test`, `manual` y `none`.** No hay tag `check`: un e2e es un test con un comando más grande, y distinguirlos no aportaba nada.

- **`test: <comando>`** es cualquier comando que salga con 0 si el claim se cumple. Delta lo ejecuta desde la raíz del repo, directo con `os/exec` y sin pasar por un shell, para que funcione igual en Windows, Mac y Linux. *Shortcut: sin shell, el comando no admite pipes, redirecciones ni `$VAR`; si hace falta algo así, se escribe como programa o como script del proyecto.* Delta no instala ni elige el runner: cada proyecto usa el suyo.
- **`manual`** (que puede llevar el paso a seguir, `manual: <paso>`) y **`none`** no tienen comando. `none` sigue siendo válido, pero queda visible en `delta status` como deuda de verificación. No hace falta montar un harness de tests el primer día.

**Reglas para que el comando no mienta** (decisión 8): el tag apunta a un archivo, paquete o directorio, **sin filtros por nombre** (`go test ./facturacion`, no `-run`), para eliminar el caso de un filtro que no coincide con nada y sale con 0. Delta no lee la salida del runner, porque ese formato es de un tercero y puede cambiar. Lo que sí hace `delta lint` es comprobar de forma estática que el ID del claim aparece en un archivo de tests (por convención de nombre: `*_test.go`, `test_*.py`, `*.test.ts`, `*.spec.ts` y similares), en el nombre del test o en un comentario. La coincidencia es tolerante a separadores y a mayúsculas (`INV-01`, `INV_01`, `INV01`, con un límite al final para que `INV01` no coincida con `INV011`), porque algunos lenguajes no admiten guiones en los nombres (`TestINV01_Redondeo` en Go).

**El archivo tiene que estar dentro de lo que el comando ejecuta.** Los IDs son únicos solo dentro de un spec, y todos los specs tienen un `INV-01`; si bastara con que cualquier test los citara, el test de `facturacion` daría `Holds` al `INV-01` de `pagos`. Por eso la cita cuenta únicamente en los archivos bajo las rutas que el comando nombra (`go test ./pagos` mira solo `pagos/`; `./...` y `.` miran todo). Un comando que no nombra ninguna ruta (`npm test`) se toma como si ejecutara todos los tests: nombrar el paquete o el archivo en el tag lo acota.

### 6.5 Seal partido en dos

- **Parte determinista** (`delta check`): ejecuta el comando de cada claim `test`. Sin modelo. El resultado depende solo del código de salida y del chequeo estático de `delta lint`:
  - sale con 0 y el ID aparece en un archivo de tests dentro de lo que el comando ejecuta: `Holds`;
  - sale con error: `Violated`;
  - sale con 0 pero ningún test cita el ID: `Unverified`, con el motivo escrito;
  - `manual` y `none`: `Unverified`.
- **Parte con LLM:** solo para claims `manual` o `none` (los `Unverified` sin comando), donde alguien tiene que leer el código y opinar. Sigue reportando, nunca corrige.

Efecto colateral buscado: escribir claims verificables se vuelve lo cómodo, porque son los que el sistema puede comprobar solo.

### 6.6 Forge explícito

Forge (antes Executor) no se disparó en 6 de 6 corridas headless. No sé la causa (*hipótesis:* depende de que el modelo elija el skill por su descripción). Lo robusto es no depender de eso:
- Entrada explícita: `/forge <slice>`.
- Si `delta status` muestra trabajo pendiente y Forge no corrió, el hook `Stop` lo señala.
- Forge puede usar `delta link` para enlazar los tests que escribe, así el lazo se cierra sin violar "nunca edita la KB".

**Ojo:** las reglas de generación de código (sección 7) se apoyan en Forge, que es la pieza más débil y menos probada. Arreglar cómo se dispara va antes que pulir las reglas.

### 6.7 Presupuesto sin romperse

El techo de 3 500 se rompe porque es un número y no una propiedad de la arquitectura. Propuesta:
- Lo siempre cargado = router + punteros + las dos reglas core. El preset `base` se carga por ruta (`paths:` sobre código fuente), no siempre. Todo lo demás, por ruta o bajo demanda (`delta context`).
- El catálogo nunca se carga por defecto; al proponer una regla se muestra su costo.
- `delta budget` falla (gate) si lo siempre cargado pasa el techo.
- Para entrar al core, una regla nueva tiene que pasar la prueba de entrada (7.2) y, si no cabe, otra sale.

---

## 7. Las reglas: cómo se organizan

Esta sección responde: *si Delta se centra en contratos verificables, ¿qué pasa con la generación de código?* El agente necesita reglas para escribir código, y algunas tienen que ser propias de Delta.

### 7.1 Tres capas

| Capa | Qué es | Dónde vive | ¿Es el centro de Delta? |
|---|---|---|---|
| **Core de Delta** | Solo dos reglas (R1 fidelidad, R2 honestidad), las que dependen del contrato y se detectan con código. Muy cortas, siempre cargadas | `.claude/rules/` sin `paths:` | Sí |
| **Aplicadas por código** | Lo que una herramienta puede decir pasa/falla (commits, lint, presupuesto, tests con ID) | Hooks, linters, `delta lint` | Sí, pero casi sin texto |
| **Preset `base`** | Calidad general al generar código: alcance (la escalera) y fiabilidad. Activado por `delta init`, desactivable | `.claude/rules/` con `paths:` sobre código fuente | No: es el beneficio secundario |
| **Presets opt-in** | Lo que depende del lenguaje o del equipo: idioms de TypeScript, DDD, seguridad, estilo | Catálogo, con `paths:` y costo visible | No: son complemento |

Las reglas de TypeScript que ya tenés **no se pierden**: las de criterio se quedan como presets con `paths: ["**/*.ts"]`; las chequeables pasan a linter.

### 7.2 Prueba de entrada al core

Una regla entra al core solo si se puede responder **qué fallo previene** y **cómo se detecta** de forma determinista. Si la detección es solo parcial o heurística, es candidata a preset (por eso alcance y fiabilidad viven en `base`) o a salir. Es la versión Delta de "cada línea le gana a la versión sin ella", y es lo que evita volver a 3 473 de 3 500 por acumulación.

### 7.3 Dimensiones de reglas para código generado

| Dimensión | Qué previene | ¿Se detecta con código? | Destino |
|---|---|---|---|
| **Fidelidad al contrato** | Código que se desvía del spec sin avisar; tests huérfanos | Sí (`delta lint`, `delta check`) | Core |
| **Alcance y tamaño** (la escalera) | Sobreingeniería, features no pedidas, ambigüedad resuelta inventando | Parcial (diff vs. alcance del slice) | Preset `base` |
| **Fiabilidad** | Errores tragados, validación omitida, lógica nueva sin test | Parcial (linter, `delta lint`) | Preset `base` |
| **Honestidad del agente** | Decir "listo" sin haber corrido nada; inventar APIs o funciones que no existen | Sí (hook `Stop`, `delta status`) | Core |
| **Cambio seguro** | Romper API pública, migraciones irreversibles, refactors oportunistas en archivos ajenos | Parcial (diff fuera del slice) | Preset |
| **Seguridad** | Input sin validar, secretos en código, authz olvidada | Parcial (escáneres, linters) | Preset con `paths:` |
| **Testabilidad** | Tiempo, random o red incrustados que impiden testear | Difícil | Preset |
| **Mantenibilidad / clean code** | Funciones enormes, duplicación, nombres opacos | Solo la parte medible | Linters + preset corto |
| **Higiene de proceso** | Commits gigantes, cambios mezclados | Sí (git hook) | Hook |

La más subestimada es **honestidad**: muchos de los fallos más dolorosos con agentes no son código feo, sino "dice que pasaron los tests y no los corrió". Es barata de imponer con un hook y tiene mucho valor.

### 7.4 El core: dos reglas, y el preset `base`

Cada regla, con el fallo que previene y cómo se detecta. El texto final conviene escribirlo corto (~300 tokens por regla).

**R1. Fidelidad: el contrato manda.**
- *Idea:* el spec del slice es la verdad y el alcance. Cada test que escribas cita el ID del claim que prueba. Si el código tiene que contradecir un claim, o la tarea no se puede decidir sin inventar, no lo tuerzas ni lo resuelvas en silencio: frená y registrá el problema como pregunta abierta del spec. (Este es el peldaño 6 de la escalera, el más Delta de todos, subido al core.)
- *Previene:* deriva silenciosa, código que contradice el contrato, tests huérfanos, ambigüedad resuelta inventando.
- *Se detecta:* `delta lint` (test sin ID, tag que apunta a un test inexistente) y `delta check` (`Violated`). Determinista.
- *Nota:* los cambios que tocan código referenciado por el spec de *otro* slice no necesitan texto; lo cubre `delta stale`.

**R2. Honestidad: nada de "listo" sin evidencia.**
- *Idea:* no se declara terminado sin haber corrido los tests y mostrar el resultado. No se inventan APIs: se verifica que existen. El cierre dice qué claims se tocaron, en qué estado quedaron y qué quedó sin verificar.
- *Previene:* el fallo "dice que pasó y no corrió", y trabajo omitido que parece terminado.
- *Se detecta:* hook `Stop` que exige que `delta check` haya corrido tras el último cambio; el estado de claims sale de `delta status`, no de la memoria del modelo.

**Preset `base`** (se activa con `delta init`, carga por `paths:` sobre código fuente). Dos reglas que no pasan la prueba de entrada del core porque su detección es parcial, pero que valen la pena por defecto:

**B1. Alcance: la escalera propia** (ver 7.5).
- *Previene:* sobreingeniería y features no pedidas.
- *Se detecta:* parcialmente, comparando el diff con los archivos y claims del slice (heurística). Sin cobertura total.

**B2. Fiabilidad: lo que nunca se recorta.**
- *Idea:* validación en límites de confianza, errores que previenen pérdida de datos, seguridad y accesibilidad no se recortan por brevedad. La lógica nueva no trivial (rama, bucle, parser, dinero, seguridad) lleva un test enlazado a su claim.
- *Previene:* un diff corto que en realidad es una regresión.
- *Se detecta:* linter (por ejemplo, catch vacío) y `delta lint` (lógica nueva sin test enlazado). Parcial.

`base` es la parte que se parece a lo que ya hacen otras herramientas. Se entrega, pero el README no lo vende: lo menciona como efecto secundario.

### 7.5 La escalera propia (B1; inspirada en ponytail, pero con peldaños de contrato)

Ponytail baja desde "¿necesita existir?" hasta "el mínimo que funciona". La de Delta conserva esa idea de reusar antes de escribir, pero arranca por el contrato y termina con la ambigüedad:

0. **Antes de empezar:** leer el código que el cambio toca y listar todo lo que debe alcanzar (callers, tests, fixtures, config, exports). Un cambio que se olvida de un caller no es más chico, es roto.
1. ¿Qué claim lo pide? Si ninguno ni ninguna petición explícita, se nombra en una línea y se omite.
2. ¿Ya existe en este slice o en otro? Se reutiliza sin duplicar.
3. ¿Lo hace la plataforma, la stdlib o una dependencia ya instalada? Se usa.
4. ¿Cabe en una línea clara? Esa línea.
5. Lo mínimo que satisface el claim, **con su test enlazado**.
6. **Si no se puede decidir sin inventar: parar y registrar la duda como pregunta abierta del spec.** Es el peldaño más Delta de todos, porque lleva la ambigüedad al contrato en vez de resolverla en silencio dentro del código. Por eso vive también en R1 (core); aquí queda como recordatorio.

### 7.6 Qué pasa con `code-lean-change`

Tu regla actual es, de hecho, la escalera de ponytail adaptada (casi los mismos peldaños, y hasta el cierre de "qué quedó sin hacer"). El archivo mismo dice que no se midió si mejora resultados, y cuesta ~922 tokens siempre cargados. Mi propuesta es no mantenerla como regla única, sino repartir lo valioso:

| Parte de `code-lean-change` | A dónde va |
|---|---|
| "Leer el código y listar lo que debe alcanzar" | B1 (`base`), peldaño 0 |
| "Cuando hay spec, sus claims son el alcance" | R1 y B1 (es lo más Delta de la regla) |
| La escalera | B1, reescrita con los peldaños de 7.5 |
| "Nunca recortar validación, errores, seguridad, accesibilidad" | B2 (`base`) |
| "Un test para lógica nueva no trivial" | B2, ahora enlazado a un claim |
| "Cerrar diciendo qué se omitió o quedó sin verificar" | R2, ahora con estado de claims |
| "Mantener la estructura y capas del código" | Preset DDD/capas |
| "Comentario `shortcut:` con el límite conocido" | Preset opcional |
| "No es objetivo tener menos líneas: el tamaño es el medio, no la meta" | Se conserva como espíritu en B1 |

**Atribución:** el sitio de ponytail indica licencia MIT. Si algo del texto original se conserva, conviene dar crédito explícito en el README o en la regla.

### 7.7 Clean code, con contrapeso

Es una etiqueta amplia y buena parte es gusto. Se trata con la misma prueba de entrada:
- Lo **medible** (complejidad, longitud de función, duplicación, imports) va a linters con hook. Casi cero texto.
- Lo **de criterio** (nombres, cohesión, límites entre módulos) va a un preset corto con `paths:`. *Hipótesis:* ahí las reglas largas rinden poco, porque el modelo ya escribe razonablemente legible por defecto.
- Lo que sí es clean code *y* es de Delta ya lo cubre la estructura: slices como bounded context y capas respetadas por el contrato.

### 7.8 Techo del core

*Propuesta:* las dos reglas core, en conjunto, pesan en el orden de ~600 tokens (≈300 cada una), y el total siempre cargado baja del techo actual al quitar `commit-messages` (~1 079 t) y `code-lean-change` (~922 t). El preset `base` (B1 + B2, ~600 t más) carga solo al tocar código fuente. Los números son una intención: se validan con `delta budget` antes de fijarlos. Si una regla nueva quiere entrar al core, otra sale o demuestra que vale más.

---

## 8. Qué adoptamos de otros (y cómo, sin volverlos el centro)

Delta no parte de cero: toma ideas de la competencia. Está bien, mientras quede claro qué se adopta y qué sigue siendo propio.

| Fuente | Qué adoptamos | Cómo se evita volverlo el centro |
|---|---|---|
| **Spec Kit / OpenSpec** | Un proceso guiado: spec → aprobar → plan → tests → código | Se mantiene en cuatro pasos, sin constitución ni puertas rígidas por fase. Si el flujo engorda, Delta se vuelve un clon de ellos. El producto es el contrato verificado, no el proceso |
| **Ponytail** | La idea de escalera (reusar antes de escribir) y la prueba "cada regla le gana a la versión sin ella" | La escalera se reescribe con peldaños de contrato (7.5) y se atribuye. La prueba de entrada (7.2) la aplica a todo el core |
| **Ponytail como complemento** | Nada: no se recomienda | Dos escaleras cargadas a la vez podrían pisarse o sumar costo, y la meta es que Delta se baste solo (decisión 6) |
| **Token Consumption Analyzer** (extensión de Spec Kit) | Medir tokens y calidad juntos al comparar configuraciones | Opcional; es de un autor individual y versión 0.1.0 |
| **Hooks de Claude Code** | Enforcement determinista para lo que no debe fallar | Es parte del diseño, no una adopción externa |

---

## 9. Cómo queda el flujo

```
delta init                                  (una vez por proyecto)
      |
      v   (por cada feature)
/quill <feature>  ("spec out <feature>")   Quill entrevista
      |                                     delta spec write -> valida y guarda (draft)
      v
delta approve <feature>                     la persona aprueba
      |
      v
/forge <feature>                            Forge: plan -> tests -> código, regido por R1, R2 y `base`
      |                                     delta link enlaza tests a claims
      v                                     hooks: lint + marcado stale
/seal <feature>                             delta check ejecuta lo verificable; LLM solo para lo manual
      |
      v
verified   (si el código cambia después -> stale, avisado solo)
```

---

## 10. Qué NO se vuelve determinista (para no engañarnos)

- Que el agente siga el plan.
- Que el agente cumpla las reglas core: los hooks y `delta lint` detectan incumplimientos, pero no garantizan que el modelo las siga a la primera.
- La calidad de los tests que escribe. Un test que pasa no es un buen test, y un test que cita un ID no prueba necesariamente el claim.
- Que Forge se dispare cuando no se lo invoca explícitamente.
- La calidad de los claims que redacta Quill.
- La honestidad: el hook exige evidencia de que se corrieron los checks, pero no que el agente interprete bien el resultado.
- Que un `Holds` pruebe que el test del claim se ejecutó: prueba que el comando salió con 0 y que existe un test, bajo las rutas del comando, que cita el ID, no que ese test corrió (un `skip` pasaría, y un comando que no nombra rutas acepta cualquier test). `delta lint` avisa de `skip`/`only`, pero es una heurística. *Mejora futura si aparece un falso `Holds` real:* un archivo de evidencia, donde cada test avisa a Delta (variable de entorno más un helper de una línea por lenguaje), que depende solo de un protocolo propio.
- Que el agente no debilite un test para que pase. Se ve en el diff, pero ninguna regla de ruta lo impide, porque cada proyecto ubica sus tests donde quiere.
- La detección de deriva por archivos referenciados es **heurística**: puede marcar `stale` de más (cambio irrelevante) o no marcarlo de menos (cambio de comportamiento sin tocar los archivos nombrados).

Los gates detectan problemas; no garantizan calidad. La promesa correcta es "los problemas no pasan desapercibidos", no "no hay problemas".

---

## 11. Decisiones abiertas

1. ~~**Orden de principios**~~ **Resuelta:** se mantiene fiabilidad > claridad > adopción > costo.
2. ~~**Orientación inicial** (lo que hacía Surveyor)~~ **Resuelta:** `delta init` genera un esqueleto determinista y un `CLAUDE.md` mínimo; los docs curados nacen a demanda cuando Quill los necesita. No hay skill de escaneo.
3. ~~**Las cuatro reglas core**~~ **Resuelta:** el core son R1 (fidelidad) y R2 (honestidad); alcance y fiabilidad pasan al preset `base`, que carga por ruta. La de "cambio seguro" sigue fuera. Queda por confirmar solo el texto final.
4. ~~**Idioma**~~ **Resuelta:** las reglas, skills y el repo en inglés; estos documentos y la conversación en español.
5. ~~**Techo del core**~~ **Resuelta:** ~600 tokens (core) y ~600 más (`base`) como estimación; se ajusta con `delta budget` sobre la marcha.
6. ~~**Ponytail como complemento**~~ **Resuelta:** no se recomienda. La meta es que Delta sea suficientemente bueno por sí solo.
7. ~~**Granularidad del `stale`**~~ **Resuelta:** por archivo referenciado, por simplicidad. Se afina si genera ruido.
8. ~~**Dónde viven los checks ejecutables**~~ **Resuelta por eliminación:** desaparece el tag `check`. Un e2e es un `test` con un comando más grande, y cada proyecto ubica sus tests donde quiera. El comando se ejecuta sin shell y sin filtros por nombre, y `delta check` no lee la salida del runner: `Holds` exige exit 0 más un test que cite el ID (6.4 y 6.5). Mejora futura anotada en §10.
9. ~~**Nombre y versión del formato de spec**~~ **Resuelta:** se mantiene el nombre actual y el formato se versiona en el frontmatter (`format: 1`), para que el CLI pueda migrar specs viejos.
10. ~~**Medir o no medir**~~ **Resuelta:** no se mide nada por ahora. El experimento viejo (~3 USD, 30 corridas) se descarta, y la medición de fidelidad de la sección 13 queda pospuesta.

---

## 12. Plan por fases

**Fase 0: limpieza (poco esfuerzo)**
- Quitar Cursor del About y del README; quitar `ROADMAP.md` plantilla.
- Poner `paths:` en las rules existentes.
- Reescribir el README con la tesis nueva y tres roles.
- Renombrar los skills (Specifier → Quill, Executor → Forge, Sentinel → Seal): carpetas, `description`, README y toda referencia cruzada.

**Fase 1: núcleo del CLI**
- Formato de spec con IDs y tags; parser EARS; `spec lint`, `spec write`, `link`, `lint`, `check`, `status`.
- Tests unitarios del CLI (gratis, sin API). Esto ya es fiabilidad demostrable.
- Suite de compatibilidad de `delta check`: proyectos mínimos reales (Go y TypeScript) con un test que pasa, uno que falla, un comando que sale con 0 sin test que cite el ID y un `skip`, para comprobar que los resultados son los de 6.5.

**Fase 2: hooks**
- `PreToolUse` que protege la KB, `PreToolUse` sobre `git commit`, `.githooks/commit-msg` con `core.hooksPath` fijado por `delta init`, aviso de `SessionStart` si falta, `delta commit-msg` en CI, `budget` como gate, `PostToolUse` con linter, `Stop` que exige evidencia.

**Fase 3: estados y Forge**
- Máquina de estados, `approve`, entrada explícita para Forge (`/forge`), `delta link` desde Forge. Es prerrequisito de las reglas de generación.

**Fase 4: reglas core y preset `base`**
- Escribir R1 y R2 con su texto final y B1 y B2 como preset `base`, medir su peso con `delta budget`, mover lo valioso de `code-lean-change` según 7.6, convertir el resto del catálogo en presets con `paths:`.

**Fase 5: deriva automática**
- `delta stale` y su integración con `Stop`. Es la pieza más diferenciadora, y también la más heurística: conviene hacerla después de tener lo anterior firme.

**Fase 6: dogfooding en Kelsen**
- Usar Delta en varias features reales. Anotar dónde estorba y dónde salva.

---

## 13. Cómo sabremos que funcionó

Sin benchmark de modelo obligatorio. Criterios que se pueden comprobar:

- El CLI tiene tests y pasan en CI.
- En Kelsen, tras varias features: ningún spec quedó desfasado *sin que el sistema lo avisara*.
- La proporción de claims en `none` baja con el tiempo (o cada `none` tiene una razón escrita).
- Los commits cumplen el formato el 100% de las veces (ahora es un hook).
- Lo siempre cargado respeta el techo sin sacar reglas a la fuerza.
- Alguien ajeno al proyecto puede hacer `delta init` y entender el flujo sin ayuda (adopción).

**Medición opcional y barata de "resultado" (fidelidad al contrato). Pospuesta por ahora (decisión 10); queda anotada por si se retoma:** construir los mismos slices con y sin las reglas core y contar, con `delta check`, cuántos claims quedan `Violated` o `Unverified`. Es determinista, no necesita juez LLM y mide justo lo que Delta promete. No sustituye un benchmark de calidad general, pero permite afirmar "fidelidad al contrato" con un número. Convendría repetir cada corrida varias veces (el agente no es determinista) y cuidar que el baseline no quede contaminado por los hooks de Delta, que fue justo el error que ponytail documentó en el suyo.

Los umbrales numéricos concretos los definís vos al arrancar la Fase 6; es mejor fijarlos *antes* de medir, no después.

---

## 14. Riesgos

| Riesgo | Mitigación |
|---|---|
| Convertirlo en algo pesado (el CLI y las reglas crecen sin control) | Pilar 9 y prueba de entrada: cada comando y cada regla tiene que ganarse su lugar; empezar con los mínimos |
| Las reglas de generación dependen de Forge, la pieza más débil | Fase 3 (Forge explícito) antes de la Fase 4 (reglas core) |
| Un clon sin `core.hooksPath` queda sin git hook y nadie lo nota | `delta init` lo fija; `SessionStart` y `delta status` avisan si falta; CI valida igual |
| El core vuelve a inflarse hasta romper el presupuesto | Techo fijado de antemano y gate con `delta budget`; si entra una, sale otra |
| Reglas de clean code largas que rinden poco | Lo medible a linters; lo de criterio, preset corto y opcional |
| Falsos positivos de `stale` que cansan | Empezar por archivo, medir ruido en Kelsen, afinar |
| Hooks que estorban al desarrollador | Bloquean al agente, no a la persona; avisos antes que bloqueos donde se pueda |
| Specs que nadie quiere escribir | `delta spec new` + Quill hacen el trabajo; los claims en `none` son válidos pero visibles |
| Sobre-prometer | Mantener separadas fiabilidad del sistema y fiabilidad del resultado (sección 3) |
| Parecerse demasiado a ponytail o a Spec Kit | Atribuir lo adoptado, mantener la escalera con peldaños de contrato y el flujo en cuatro pasos (sección 8) |
| Dependencia de la plataforma (hooks de Claude Code cambian) | El núcleo vive en el CLI; los hooks son un adaptador pequeño |

---

## 15. Posicionamiento (cómo explicarlo a otros)

| Herramienta | Qué ofrece | Qué no cubre |
|---|---|---|
| Spec Kit / OpenSpec | Proceso guiado por fases para cambiar código | No detecta que el contrato quedó viejo |
| Ponytail | Reglas que reducen el código escrito, con benchmark público | No se ocupa del conocimiento del proyecto ni del contrato |
| **Delta (nuevo rumbo)** | **Contratos vivos verificables, código generado fiel al contrato y deriva detectada automáticamente** | No promete mejor código en general ni ahorro de tokens (el preset `base` aporta código más limpio como efecto secundario, sin garantía) |

---

## 16. Glosario

- **Slice:** una capacidad acotada del proyecto (como un bounded context), con su propio contrato.
- **Contrato vivo (spec):** archivo que dice qué es verdad ahora; se edita en sitio y no acumula historia.
- **Claim:** una afirmación del contrato (invariante o criterio de aceptación) con ID y tag.
- **EARS:** notación para escribir requisitos con estructura fija ("Cuando X, el sistema debe Y").
- **Tag de verificación:** `test: <comando>`, `manual` o `none`; dice cómo se comprueba el claim.
- **Deriva (drift):** diferencia entre lo que dice el contrato y lo que hace el código.
- **Stale:** slice cuyo contrato o código cambió sin que se haya vuelto a verificar.
- **Regla core:** regla corta y siempre cargada sobre la relación entre código y contrato (hoy R1 fidelidad y R2 honestidad).
- **Preset `base`:** reglas de calidad general al generar código (alcance y fiabilidad). Se activa con `delta init`, carga por ruta y es el beneficio secundario de Delta.
- **Quill, Forge, Seal:** los tres skills, que redactan, construyen y certifican el contrato (antes Specifier, Executor y Sentinel).
- **`core.hooksPath`:** ajuste local de git que apunta a una carpeta versionada (`.githooks/`) para compartir hooks entre clones; lo fija `delta init`.
- **Preset:** conjunto opcional de reglas (lenguaje, DDD, seguridad) que se activa por proyecto y por ruta.
- **Prueba de entrada:** pregunta "qué fallo previene y cómo se detecta" que decide si una regla entra al core.
- **Hook:** comando que Claude Code ejecuta en un evento (antes de editar, al terminar, etc.); a diferencia de una instrucción en `.md`, se ejecuta siempre.
- **Harness:** el andamiaje alrededor del modelo (herramientas, permisos, reglas, hooks); en Claude Code se configura, no se construye desde cero.

---

## Fuentes consultadas

- Ponytail: https://ponytail.dev/ y su README de benchmark agentic
- Token Consumption Analyzer para Spec Kit: https://speckit-community.github.io/extensions/token-analyzer
- Críticas de consumo de tokens en Spec Kit: https://news.leodex.io/news/github-spec-kit-spec-first-approach-reshapes-ai-coding-economics
- Comparativa OpenSpec vs Spec Kit: https://codemyspec.com/blog/openspec-vs-spec-kit
- Hooks en Claude Code: https://blakecrosley.com/zh-Hans/blog/claude-code-hooks-explained
- Hooks vs CLAUDE.md: https://dev.to/gentic_news/stop-relying-on-claudemd-for-guarantees-build-deterministic-hooks-instead-35kp
- Rules con `paths:`: https://www.dsebastien.net/claude-code-memory/
- Repo de Delta: https://github.com/ronz204/deltas-ai