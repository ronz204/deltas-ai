# Delta: roadmap

*31 de octubre de 2026. Estado actualizado el 1 de noviembre de 2026. Orden de trabajo para llevar a Delta al rumbo de `redirection.md`.*

Este documento ordena el trabajo: qué se hace, en qué orden, por qué ese orden, cuándo una fase se da por cerrada y **cómo va cada fase**. El diseño y las decisiones están en `redirection.md` (las secciones se citan como §N); el contexto, en `background.md`; la parte determinista, en `deterministic.md`. Si algo choca, manda `redirection.md`.

Todas las decisiones abiertas de §11 están resueltas, así que nada de lo que sigue espera una respuesta previa.

---

## Estado

**Avance total: 14 de 64 tareas, 22 %**

| Fase | Nombre | Estado | Avance |
|---|---|---|---|
| 0 | Limpieza y renombrado | ✅ Terminada | 11/11, 100 % |
| 1 | Núcleo del CLI | 🔄 En progreso | 3/11, 27 % |
| 2 | Hooks | 🔲 Pendiente | 0/11, 0 % |
| 3 | Estados y Forge | 🔲 Pendiente | 0/9, 0 % |
| 4 | Reglas core y preset `base` | 🔲 Pendiente | 0/9, 0 % |
| 5 | Deriva automática | 🔲 Pendiente | 0/6, 0 % |
| 6 | Dogfooding en Kelsen | 🔲 Pendiente | 0/7, 0 % |

Leyenda: 🔲 Pendiente · 🔄 En progreso · ✅ Terminada.

**Cómo se mantiene.** Cada fase cuenta como tarea cada casilla de sus entregables y de sus criterios de cierre. El porcentaje es siempre casillas marcadas entre casillas totales, y se recalcula cada vez que se marca o se añade una. Una fase solo pasa a ✅ cuando no le queda ninguna casilla sin marcar; si algo no se puede cerrar, se divide la tarea en vez de redondear. Si se añade una tarea, también se suma al total de arriba.

---

## Orden y dependencias

```
Fase 0  limpieza y renombrado
   |
Fase 1  núcleo del CLI  ------------------+
   |                                       |
Fase 2  hooks                              |
   |                                       |
Fase 3  estados y Forge                    |
   |                                       |
Fase 4  reglas core y preset base          |
   |                                       |
Fase 5  deriva automática (delta stale) <--+
   |
Fase 6  dogfooding en Kelsen
```

Por qué este orden:

- **La Fase 1 va antes que todo lo demás** porque `lint`, `check` y `status` son la base de los hooks, de los estados y de las reglas. Sin ellos, los hooks no tienen nada que invocar.
- **La Fase 3 va antes que la 4** porque R1 y R2 se apoyan en Forge, que es la pieza más débil (no se disparó en 6 de 6 corridas). Pulir reglas sobre un skill que no arranca no sirve.
- **La Fase 5 va al final de la construcción** porque es la más diferenciadora y también la más heurística; conviene tener lo anterior firme antes de medir su ruido.
- **La Fase 6 es la única que valida el conjunto.** No hay benchmark de modelo (decisión 10): el criterio es usarlo en un repo real.

Regla de trabajo: cada fase termina con sus criterios cumplidos y con un commit (o varios, uno por concern) antes de abrir la siguiente.

---

## Fase 0: limpieza y renombrado

**Estado: ✅ Terminada, 11/11.**

*Poco esfuerzo, sin código. Deja el repo coherente con la nueva dirección.*

**Entregables**
- [x] Quitar Cursor del README (no había menciones; se verificó con una búsqueda).
- [x] Quitar Cursor del About del repo. Es un ajuste de GitHub, no un archivo, y es visible públicamente, así que lo hace Roney o se confirma el texto antes.
- [x] Quitar la plantilla `ROADMAP.md` de la raíz (ningún skill la lee).
- [x] Crear `harness/` como copia ya renombrada de lo que hoy vive en `.claude/` (Specifier → Quill, Executor → Forge, Sentinel → Seal): carpetas, `description` y referencias cruzadas entre skills, rules y docs. `.claude/` no se toca: contiene las reglas que rigen el trabajo en este repo y se regenera desde `harness/` cuando exista `delta init` (Fase 2). *`archivist` y `surveyor` van en la copia de forma transitoria hasta que el CLI los reemplace.*
- [x] Reorganizar `harness/` con `core/{configs,githooks,templates}` para los archivos de soporte (`deterministic.md`, §3 y §12). Hoy solo existe `core/configs/` (`settings.json` y `budget.conf`); `templates/` llega en la Fase 1 y `githooks/` en la Fase 2.
- [x] Poner `paths:` en las rules existentes (§7.1). *Tres ya lo tenían; se añadió a `code-lean-change`. `commit-messages` queda global a propósito: un hook la reemplaza.*
- [x] Reescribir el README con la tesis nueva y tres roles, mencionando el preset `base` solo como efecto secundario (§7.4).
- [x] Ampliar `.gitattributes` para fijar LF también en `*.go`, `go.mod`, `go.sum` y las carpetas de git hooks. `*.md` (incluidos los de `elements/`) ya estaba cubierto.
- [x] Commit(s) de la fase, uno por concern.

**Cierra cuando**
- [x] Una búsqueda de `Specifier`, `Executor`, `Sentinel` y `Cursor` no devuelve resultados fuera de `.claude/` (que se regenera en la Fase 2) y de los documentos de `elements/` que los citan como historia. *Salvo el README, que los cita a propósito al describir la generación anterior.*
- [x] Cada skill se llama por su nombre nuevo y su `description` nombra el verbo (redactar, construir, verificar), no solo la metáfora.

**Cuidado:** el disparo de un skill depende de su `description`, no de su nombre. Revisar que renombrar no la degrade. *Las tres descripciones ya nombraban el verbo y no se tocaron.*

---

## Fase 1: núcleo del CLI

**Estado: 🔄 En progreso, 3/11.**

*El proyecto Go y la parte determinista. Ya es fiabilidad demostrable sin tocar a Claude Code.*

**Entregables** (§6.1, §6.4, §6.5; estructura y orden de implementación en `deterministic.md`)
- [x] Esqueleto del proyecto Go: `go.mod` en la raíz, `entries/`, `internal/` (capas `domain`, `usecase`, `gateway`, `adapter`) y `harness/` con su `harness.go` y `core/templates/` (spec y `CLAUDE.md`), más el test que hace cumplir la regla de dependencia entre capas. *Hoy solo existe la capa `domain`; las otras tres nacen con su primer paquete, y el test ya las cubre.*
- [x] Formato de spec con `format: 1`, IDs por claim y tags `test: <comando>`, `manual`, `none`. *`domain/{diag,slice,claim,spec}`, con fuzzing del parser.*
- [x] Parser EARS y de Given/When/Then. *En `domain/claim`; se ejecuta dentro de `spec.Parse`.*
- [ ] Comandos: `delta spec new`, `spec lint`, `spec write`, `link`, `lint`, `check`, `status`, `commit-msg`.
- [ ] `delta check` como en §6.5: ejecución directa con `os/exec` (sin shell), `Holds` solo con exit 0 más un test que cita el ID, `Unverified` si falta el test, `Violated` si falla. No lee la salida del runner.
- [ ] `delta lint`: coincidencia tolerante del ID (`INV-01`, `INV_01`, `INV01`), aviso de `skip` y `only`.
- [ ] Tests unitarios del CLI.
- [ ] Suite de compatibilidad de `delta check`: proyectos mínimos reales en Go y TypeScript con un test que pasa, uno que falla, un comando que sale con 0 sin test que cite el ID y un `skip`.

**Cierra cuando**
- [ ] Los tests del CLI y la suite de compatibilidad pasan en CI.
- [ ] Se puede crear un spec, enlazarlo a un test real y obtener `Holds`, `Violated` y `Unverified` en los casos esperados, sin ningún LLM.
- [ ] Commit(s) de la fase, uno por concern.

**Aún sin hacer aquí:** `approve`, estados, `context`, `budget` y `stale` (fases 2, 3 y 5).

---

## Fase 2: hooks

**Estado: 🔲 Pendiente, 0/11.**

*Volver obligatorio lo que hoy es consejo (§6.2).*

**Entregables**
- [ ] `delta hook <evento>`: el adaptador que lee el JSON de cada evento y decide con `usecase/guard`. Antes de escribirlo, verificar el formato real del JSON en la documentación de hooks vigente.
- [ ] Hooks de Claude Code en `.claude/settings.json`: `SessionStart`, `PreToolUse` sobre la KB, `PreToolUse` sobre `git commit`, `PostToolUse` con el linter, `Stop` que exige evidencia de `delta check`. Todos bloquean con exit **2**, no 1.
- [ ] `harness/core/githooks/commit-msg`, que `delta init` instala como `.githooks/commit-msg` (script POSIX que llama a `delta commit-msg`) y `core.hooksPath` fijado por `delta init`, sin pisar uno ajeno.
- [ ] Aviso en `SessionStart` y `delta status` si falta `core.hooksPath` o el binario `delta`.
- [ ] `delta commit-msg` en CI sobre el rango del PR.
- [ ] `delta budget` como gate.
- [ ] `delta init` con defaults sensatos (esqueleto determinista, `CLAUDE.md` mínimo), que además regenera `.claude/` desde `harness/`.

**Cierra cuando**
- [ ] Un commit del agente con formato inválido es bloqueado, y uno de una persona también (con el hook instalado).
- [ ] Un clon sin `core.hooksPath` lo avisa al iniciar sesión.
- [ ] Un hook que falla porque falta `delta` avisa y no deja la sesión bloqueada.
- [ ] Commit(s) de la fase, uno por concern.

---

## Fase 3: estados y Forge

**Estado: 🔲 Pendiente, 0/9.**

*Prerrequisito de las reglas de generación (§6.3, §6.6).*

**Entregables**
- [ ] Máquina de estados `draft → approved → built → verified`, con `stale` como retorno, guardada en el frontmatter y modificada solo por el CLI. Decidir aquí qué comando dispara `approved → built` y `built → verified` (`deterministic.md`, §16).
- [ ] `delta approve <slice>` (lo ejecuta la persona).
- [ ] Forge con entrada explícita (`/forge <slice>`) y sin arrancar si el spec no está `approved` y limpio.
- [ ] Forge usa `delta link` para enlazar los tests que escribe, sin editar la KB.
- [ ] Quill y Seal adaptados: Quill escribe vía `delta spec write`; Seal usa `delta check` y solo usa el LLM para `manual` y `none`. Con eso salen `archivist` y `surveyor` de `harness/`.
- [ ] El hook `Stop` señala el trabajo pendiente si Forge no corrió.

**Cierra cuando**
- [ ] Forge se dispara por `/forge` en una corrida headless de prueba. (El 0 de 6 del piloto era el síntoma; este es el criterio.)
- [ ] El ciclo completo funciona de punta a punta en un slice de prueba: Quill → `approve` → Forge → Seal.
- [ ] Commit(s) de la fase, uno por concern.

---

## Fase 4: reglas core y preset `base`

**Estado: 🔲 Pendiente, 0/9.**

*Escribir el texto final, con el peso medido (§7).*

**Entregables**
- [ ] R1 (fidelidad) y R2 (honestidad) como reglas core, ~600 tokens en conjunto.
- [ ] B1 (alcance, la escalera) y B2 (fiabilidad) como preset `base`, con `paths:` sobre código fuente, activado por `delta init` y desactivable.
- [ ] Mover lo valioso de `code-lean-change` según §7.6 y borrarla como regla única; dar crédito a ponytail (licencia MIT) donde se conserve texto.
- [ ] Eliminar la regla `commit-messages` como texto (ya la cumple el hook).
- [ ] Convertir el resto del catálogo en presets con `paths:`; mover lo chequeable a linters.
- [ ] Reglas, skills y repo en inglés (decisión 4).

**Cierra cuando**
- [ ] `delta budget` confirma que lo siempre cargado baja del techo anterior y deja margen. Las cifras de §7.8 son una estimación; aquí se ajustan con medición real.
- [ ] Cada regla del core responde a "qué fallo previene y cómo se detecta" (prueba de entrada, §7.2).
- [ ] Commit(s) de la fase, uno por concern.

---

## Fase 5: deriva automática

**Estado: 🔲 Pendiente, 0/6.**

*La pieza más diferenciadora y la más heurística (§6.1, §10).*

**Entregables**
- [ ] `delta stale`: compara los archivos modificados con los que cada spec referencia, por archivo (decisión 7).
- [ ] Integración con `PostToolUse` y `Stop`.
- [ ] `delta status` muestra `stale` por slice.

**Cierra cuando**
- [ ] Cambiar un archivo referenciado devuelve el slice a `stale` y se avisa solo.
- [ ] Se conocen los falsos positivos más frecuentes en un caso real, para decidir si hay que afinar la granularidad. Esta fase se cierra de verdad en la Fase 6.
- [ ] Commit(s) de la fase, uno por concern.

---

## Fase 6: dogfooding en Kelsen

**Estado: 🔲 Pendiente, 0/7.**

*La validación. Sin benchmark ni medición de modelo (decisión 10).*

**Qué se hace**
- [ ] Usar Delta en varias features reales de Kelsen, de punta a punta.
- [ ] Anotar dónde estorba y dónde salva.
- [ ] Fijar *antes* de empezar los umbrales de §13, para no ajustarlos después de ver los resultados.

**Cierra cuando** (§13)
- [ ] Ningún spec quedó desfasado sin que el sistema lo avisara.
- [ ] La proporción de claims en `none` baja, o cada `none` tiene una razón escrita.
- [ ] Los commits cumplen el formato siempre.
- [ ] Alguien ajeno al proyecto puede hacer `delta init` y seguir el flujo sin ayuda.

**Qué sale de aquí:** una lista de ajustes (ruido de `stale`, reglas que sobran o faltan, fricción de adopción), no una promesa de resultado.

---

## Cosas anotadas para más adelante

No están en ninguna fase. Cada una tiene que ganarse su lugar con la prueba de entrada (pilar 9) antes de entrar:

- **Archivo de evidencia** para `delta check`: cada test avisa a Delta con un helper de una línea por lenguaje. Solo si aparece un falso `Holds` real (§10).
- **Checks integrados en el CLI** (por ejemplo, "este paquete no importa aquel") para no escribir scripts.
- **Reportes JUnit** como evidencia, si el ruido del chequeo estático se vuelve un problema.
- **Medición de fidelidad al contrato** con y sin las reglas core (§13), hoy pospuesta.
- **Granularidad más fina de `stale`** (directorio o símbolo), si por archivo genera ruido.
- **Preset de "cambio seguro"**, hoy fuera del core.

---

## Riesgos que afectan al orden

| Riesgo | Qué hacer |
|---|---|
| Forge sigue sin dispararse al final de la Fase 3 | Detener la Fase 4. Las reglas de generación dependen de él (§14) |
| El CLI crece más de lo previsto | Cada comando nuevo pasa la prueba de entrada; si no se puede justificar, queda en "para más adelante" |
| Un falso `Holds` real aparece antes de la Fase 6 | Subir el archivo de evidencia de "anotado" a la Fase 1 o 2 |
| Los hooks de Claude Code cambian | El núcleo vive en el CLI; solo se reescribe el adaptador (§14) |
