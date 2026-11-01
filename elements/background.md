# Delta: background

*Contexto y conceptos, 30 de octubre de 2026.*

Este documento es el **telón de fondo** de `redirection.md`. La división es esta:

- **`redirection.md`:** las decisiones, el diseño, el plan y los números del piloto y de la investigación (su sección 2).
- **`background.md` (este):** lo que ayuda a entender esas decisiones y que allá no cabe: cómo funciona Delta hoy, conceptos de Claude Code, cómo se mide un harness, detalle de los referentes y cómo llegamos hasta acá.

Leelo antes de `redirection.md`, o pasáselo a Claude Code como contexto. Lo que es dato, propuesta o hipótesis está marcado.

**Límite:** no pude leer la carpeta `.claude/` del repo (el sitio bloquea el acceso automatizado). Trabajé con el README, `summary.md` y `code-lean-change.md`. Lo que digo de los skills y scripts internos viene de esos documentos, no de haberlos leído.

---

## 1. Delta hoy, en corto

**Qué es:** una metodología de documentación viva para agentes de código, empaquetada como harness de Claude Code (`.claude/`). El repo (https://github.com/ronz204/deltas-ai) no tiene código de aplicación: el producto es el harness.

**Idea central:** el contexto de un proyecto no debería vivir en la cabeza de alguien ni en una bitácora que cuesta más leer de lo que vale. Se guarda como **contratos vivos**: archivos que dicen qué es verdad *ahora*, se editan en sitio y nunca acumulan historial (git ya guarda el pasado).

**Objetivo original:** algo ligero, fácil de usar y adoptar, con claridad para el agente y para el desarrollador. Es open source. La preocupación por tokens, punteros de conocimiento y carga bajo demanda llegó después.

**Los cinco skills:** Surveyor (arranca un repo sin KB), Specifier (entrevista una feature), Archivist (único escritor de la KB), Sentinel (compara spec contra código y reporta deriva) y Executor (plan aprobado, tests primero, luego código).

**La KB, en tres capas:** `CLAUDE.md` como router; `.claude/docs/` como referencia curada (solo lo que el código no muestra); `deltas/<slice>.spec.md` como contrato por feature, con claims en EARS y una etiqueta de verificación (`test`, `check`, `manual`, `none`).

**Flujo:** copiar `.claude/` → "bootstrap this repo with Delta" → por feature: "spec out X" → "implement X" (el plan se aprueba antes) → "check X for drift".

**Lo que no hace:** no lleva registro de decisiones; Sentinel solo verifica specs y designs; Executor no toca la KB; el spec no se mantiene solo (alguien tiene que pedir la actualización); no corre fuera de Claude Code; nunca se probó de punta a punta en un proyecto grande.

**Detalle pendiente:** el About del repo todavía dice "Claude Code & Cursor".

## 2. Claude Code y su harness

**Es un agente:** un loop donde el modelo decide qué herramienta usar (leer, editar, correr comandos, buscar), ve el resultado y sigue hasta terminar. Actúa sobre el proyecto, no solo responde.

**A Claude Code no se le construye un harness: se configura el que trae.** Si se quiere uno propio desde cero, existe el Agent SDK, que da el mismo loop como librería.

| Componente | Para qué sirve |
|---|---|
| `CLAUDE.md` | Instrucciones persistentes del proyecto, cargadas cada sesión |
| Rules (`.claude/rules/`) | Instrucciones por tema; con `paths:` en el frontmatter cargan solo al tocar archivos que coinciden |
| Skills | Carpetas de instrucciones (y scripts) que el agente carga cuando la tarea lo amerita |
| Subagentes | Agentes especializados con contexto y herramientas propios |
| Hooks | Comandos que se ejecutan en eventos (antes/después de una herramienta, al iniciar, al terminar) |
| Slash commands, MCP, plugins | Empaquetar flujos y conectar herramientas externas |
| Permisos y settings | Qué puede hacer el agente sin preguntar |

La distinción clave (instrucciones probabilísticas vs. hooks deterministas, y el detalle del exit 2) está en `redirection.md`, sección 2.

## 3. Cómo se mide un harness

Pregunta de partida: ¿cómo se mide el rendimiento de un agente con harness configurado?

1. **Un set fijo de tareas representativas**, idealmente de tu propio repo, congeladas en un commit. Los benchmarks genéricos (tipo SWE-bench) miden modelo + andamiaje, no tu configuración en tu proyecto.
2. **Comparar A vs. B cambiando una sola cosa.** Cambiar tres a la vez no dice qué ayudó.
3. **Repetir cada corrida varias veces:** el agente no es determinista.
4. **Métricas verificables sin opinar:** tests que pasan, regresiones, tokens/costo/tiempo, intervenciones manuales, cuánto del diff se rehízo, adherencia a reglas (por lint, grep o hook).
5. **LLM-as-judge solo para lo difuso**, con sus sesgos presentes.
6. **En la práctica:** modo headless (`claude -p`), un worktree por corrida y revisión de transcripts (explican *por qué* falló, a veces más que el número).
7. **Cuidar el baseline.** Ponytail documentó que el suyo estuvo contaminado porque el hook del plugin corría también en el brazo "sin skill".

**Diagnóstico del primer benchmark de Delta** (aprendizaje, no número): no estaba fijado el objetivo, así que no se sabía qué medir; el fixture era de ~150 líneas, con nada que explorar; la tarea ya saturaba la corrección; y se midió costo sin una hipótesis detrás. Antes hubo también un benchmark de Sentinel que se construyó y se purgó por costo y saturación.

## 4. El trade-off de fondo

Toda decisión de diseño cuesta algo. Orientar mejor al agente cuesta tokens fijos (~4.2k por sesión en el piloto): **el costo es seguro, el ahorro es una hipótesis.**

El punto de equilibrio es simple: Delta se paga si ahorra más de esos ~4.2k tokens en exploración por sesión. Es plausible donde el agente lee 20-40 archivos, pero nadie lo midió. Es exactamente donde Delta apuesta (repos grandes, trabajo repetido, documentación que sobrevive entre sesiones), y lo que el piloto no pudo ver. Los números del piloto están en `redirection.md`, sección 2.

## 5. Referentes: el detalle que no está en `redirection.md`

**Spec Kit** es *spec-first*: la especificación guía el diseño inicial, repartida en varios archivos por feature, con fases en secuencia. Sus pasos de clarify/analyze son consultivos. GitHub lo describe como experimental. Hay una extensión comunitaria (*Token Consumption Analyzer*, versión 0.1.0) para medir tokens y calidad entre flujos; su reporte de un preset "lean" (−46.8% tokens, −0.07 de calidad) viene de un autor individual.

**OpenSpec** es más liviano y orientado a código existente. Un comparativo de terceros cita ~250 líneas de salida contra ~800 de Spec Kit. Su verificación tampoco bloquea. En ambos, el spec gobierna el código *por convención, no por enforcement*, y nada cierra el lazo entre spec y software corriendo: ese es el hueco que Delta puede ocupar.

**Ponytail** (MIT, según su sitio) es un ruleset de comportamiento: su mecanismo es leer el código primero y subir una escalera (¿necesita existir?, ¿ya está en el código?, ¿lo hace la stdlib o la plataforma?, ¿una dependencia instalada?, ¿una línea?, el mínimo). Benchmark publicado por el propio autor: Opus 5.5, 39 tareas, 5 corridas cada una, checks ocultos de corrección y seguridad, rivales incluidos (entre ellos un prompt de una línea, "Follow YAGNI principles"). Reporta −53% líneas, −45% tokens de salida, −26% costo, −41% tiempo, con corrección similar. Sirven como referencia de *método*. El ahorro es sobre todo de *salida*; el costo total baja menos porque leer contexto sigue costando igual.

**Otras ideas vistas de pasada** (sin verificar a fondo): GSD, que ataca el "context rot" ejecutando tareas chicas en contextos nuevos; LeanSpec, que propone specs dentro de ~2 000 tokens (afirmación del propio proyecto).

## 6. Cómo llegamos hasta acá (29-30 de octubre de 2026)

1. Aclaramos el modelo mental: Claude Code es un agente; se configura su harness, no se construye uno.
2. Vimos que Spec Kit, OpenSpec y ponytail imponen proceso o comportamiento y casi no publican evaluaciones rigurosas de mejora.
3. Diagnosticamos el benchmark fallido: faltaba un objetivo claro.
4. Aceptamos el trade-off: más orientación y flujo cuestan más tokens; el ahorro deja de ser la tesis.
5. Elegimos la tríada Specifier, Sentinel, Executor y un CLI en Go para lo determinista.
6. Cambiamos la tesis a **fiabilidad verificable** y fijamos principios con orden de prioridad.
7. Separamos rules de criterio y rules chequeables (lo chequeable va a linters y hooks).
8. Notamos que Delta adopta ideas de la competencia y decidimos no volverlas el centro.
9. Detectamos que los contratos no cubrían la **generación de código** y definimos reglas core propias, una escalera propia y presets.
10. `code-lean-change` se disuelve en esas reglas.
11. (31 de octubre) Los tres skills se renombran Quill, Forge y Seal. Los hooks de git se comparten con `core.hooksPath` más una capa propia para el agente. El core baja a dos reglas (fidelidad, honestidad) y el resto pasa al preset `base`, porque mejorar el código es un beneficio secundario de Delta. En este documento, "Specifier, Sentinel, Executor" son los nombres previos a ese cambio.

El diseño resultante de todo esto está en `redirection.md`.

## 7. Qué se sabe y qué no

| | Contenido |
|---|---|
| **Dato** | Lo medido en el piloto; lo publicado por ponytail como referencia de método; las críticas de tokens a Spec Kit; el comportamiento de hooks y rules con `paths:` |
| **Propuesta** | La tesis, el CLI y sus comandos, los estados del slice, los IDs por claim, las reglas core y la escalera propia |
| **Hipótesis** | Que "0 `test` / 14 `none`" sea estructural (nadie podía enlazar test y claim); que Executor no se dispare por depender de la descripción del skill; que las reglas largas de clean code rindan poco; que ponytail y la escalera de Delta convivan sin pisarse; que Delta compense en repos grandes |
| **Desconocido** | Contenido real de los skills y scripts; rendimiento de Delta en un repo real; cuánto ruido genera el `stale` por archivo referenciado |

## 8. Para retomar en Claude Code

Traé `executor/SKILL.md`, `budget.sh` y `links.sh`: con eso se confirman o corrigen las hipótesis sobre Executor y el lazo de tags. Después, las decisiones abiertas están en `redirection.md`, sección 11, y el plan por fases en la sección 12.

**Referencia adicional** (no está en `redirection.md`): buenas prácticas para benchmarks agentic (ABC), https://arxiv.org/html/2507.02825v4.