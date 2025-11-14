# Bsmart Backend Challenge - API de Gestión de Productos

API RESTful lista para producción con soporte de WebSocket en tiempo real para gestión de productos y categorías. Construida con Go, PostgreSQL y Redis.

---

## Deploy

**URL del despliegue**: https://bsmart-challenge.onrender.com/

> **⚠️ Nota importante**: La primera request puede tardar entre 30-60 segundos en responder, ya que el servidor entra en modo sleep después de 15 minutos de inactividad. Las siguientes peticiones funcionarán con normalidad.

---

## Documentación Swagger

La API cuenta con documentación interactiva completa. Una vez que la aplicación esté corriendo, accede a:

```
http://localhost:8080/api/docs/index.html
```

o accede a la documentación desde el deploy:

```
https://bsmart-challenge.onrender.com/api/docs/index.html
```

Desde Swagger UI podrás explorar todos los endpoints, ver ejemplos de requests/responses y probar la API directamente desde el navegador.

---

## Tabla de Contenidos

- [Deploy](#-deploy)
- [Documentación Swagger](#-documentación-swagger)
- [Características](#características)
- [Stack Tecnológico](#stack-tecnológico)
- [Arquitectura](#arquitectura)
- [Instalación Local](#instalación-local)
- [Características Clave de la Base de Datos](#características-clave-de-la-base-de-datos)
- [WebSockets - Actualizaciones en Tiempo Real](#websockets---actualizaciones-en-tiempo-real)
- [Gestión de Base de Datos](#gestión-de-base-de-datos)
- [Docker](#docker)
- [Consideraciones de Performance](#-consideraciones-de-performance)
- [Consideraciones de Seguridad](#-consideraciones-de-seguridad)
- [Documentos Adicionales](#-documentos-adicionales)
- [👥 Autor](#-autor)

---

## Características

### Funcionalidad Principal

- **Gestión de Productos**: Operaciones CRUD completas para productos con relaciones de categorías
- **Gestión de Categorías**: Sistema completo de categorías con asociaciones many-to-many de productos
- **Historial de Productos**: Seguimiento automático de cambios de precio y stock mediante triggers de PostgreSQL
- **Búsqueda Universal**: Búsqueda full-text en productos y categorías
- **Paginación y Filtrado**: Paginación personalizable con soporte de ordenamiento y filtrado

### Autenticación y Autorización

- **Autenticación JWT**: Autenticación segura basada en tokens
- **Control de Acceso Basado en Roles**: Roles Admin y Cliente con diferentes permisos
- **Rutas Protegidas**: Protección de rutas mediante middleware

### Actualizaciones en Tiempo Real

- **Soporte WebSocket**: Actualizaciones en vivo para todas las operaciones CRUD
- **Broadcasting de Eventos**: Emisión automática de eventos para cambios en productos/categorías
- **Patrón Hub-Client**: Gestión escalable de conexiones WebSocket

### Características de Base de Datos

- **Seguimiento Automático de Historial**: Trigger de PostgreSQL registra cambios de precio/stock
- **Búsqueda Full-Text**: Índice GIN en nombres de productos para búsqueda eficiente
- **Queries Optimizadas**: Índices estratégicos en campos frecuentemente consultados
- **Soporte de Transacciones**: Operaciones seguras en múltiples tablas

---

## Stack Tecnológico

### Backend

- **Go 1.25.4**: Lenguaje de programación principal
- **Gin**: Framework web HTTP de alto rendimiento
- **pgx/v5**: Driver y toolkit de PostgreSQL

### Base de Datos

- **PostgreSQL 15**: Base de datos principal con características avanzadas (triggers, índices GIN)

### Autenticación

- **JWT (golang-jwt/jwt/v5)**: Autenticación basada en tokens
- **bcrypt**: Hashing de contraseñas

### WebSocket

- **gorilla/websocket**: Implementación de WebSocket

### Herramientas de Desarrollo

- **Docker & Docker Compose**: Entorno de desarrollo containerizado
- **golang-migrate**: Gestión de migraciones de base de datos
- **godotenv**: Gestión de variables de entorno

---

## Arquitectura

### Patrón Arquitectónico: MVC

Elegí una **arquitectura simple y tipo MVC** que balancea simplicidad con mantenibilidad:

```
┌─────────────────────────────────────────────────────────────┐
│                     Petición HTTP                           │
└────────────────────────┬────────────────────────────────────┘
                         ↓
                   ┌──────────┐
                   │  Router  │
                   └─────┬────┘
                         ↓
              ┌──────────────────────┐
              │    Middleware        │
              │  • Logger            │
              │  • Error Handler     │
              │  • JWT Auth          │
              │  • Role Check        │
              └──────────┬───────────┘
                         ↓
                   ┌──────────┐
                   │ Handlers │ (Controladores)
                   │          │ • Parsear petición
                   │          │ • Validar
                   │          │ • Llamar a DB
                   │          │ • Emitir evento WS
                   │          │ • Retornar respuesta
                   └─────┬────┘
                         ↓
                   ┌──────────┐
                   │ Capa DB  │ (Queries)
                   │          │ • Operaciones CRUD
                   │          │ • Transacciones
                   └─────┬────┘
                         ↓
                   ┌──────────┐
                   │PostgreSQL│
                   └──────────┘

        (paralelo) ──→ ┌──────────────┐
                       │ WebSocket Hub│
                       │ • Broadcast  │
                       └──────────────┘
```

## Instalación Local

### Prerrequisitos

Antes de comenzar, asegúrate de tener instalado:

- **Go 1.25.4** o superior → [Descargar Go](https://golang.org/dl/)
- **Docker** y **Docker Compose** → [Descargar Docker](https://www.docker.com/get-started)
- **Git** → [Descargar Git](https://git-scm.com/downloads)

### Paso 1: Clonar el Repositorio

Abre tu terminal y ejecuta:

```bash
git clone <url-del-repositorio>
cd bsmart-backend
```

### Paso 2: Configurar Variables de Entorno

Copia el archivo de ejemplo de configuración:

```bash
# Linux / Mac
cp .env.example .env

# Windows (CMD)
copy .env.example .env

# Windows (PowerShell)
Copy-Item .env.example .env
```

```env
DATABASE_URL=postgresql://bsmart:bsmart_pass@localhost:5432/bsmart_dev?sslmode=disable
PORT=8080
JWT_SECRET=dev_secret_key_change_in_production
```

**Nota Importante**:

- ⚠️ DEBES cambiar `JWT_SECRET` por una clave segura generada aleatoriamente:
  ```bash
  openssl rand -base64 32
  ```

### Paso 3: Iniciar Servicios de Base de Datos

Inicia PostgreSQL usando Docker Compose:

```bash
docker-compose up -d db
```

**Qué hace este comando:**

- `-d`: Ejecuta los contenedores en segundo plano (detached mode)
- `db`: Inicia PostgreSQL 15 en el puerto 5432

**Verificar que los servicios están corriendo:**

```bash
docker-compose ps
```

**Output esperado:**

```
NAME                          STATUS
bsmart-backend_db_1          Up

```

**Si los servicios no están "Up"**, revisa los logs:

```bash
docker-compose logs db
```

### Paso 4: Ejecutar Migraciones de Base de Datos

Aplica el schema de la base de datos:

```bash
docker-compose run --rm migrate up
```

**Qué hace este comando:**

- `run --rm`: Ejecuta un contenedor temporal que se elimina al terminar
- `migrate up`: Aplica todas las migraciones pendientes

**Output esperado:**

```
Applying migration 000001_first_migration.up.sql
Migration complete
```

**Solución de problemas:**

- Si dice "database is dirty", ejecuta: `docker-compose run --rm migrate force 000001`
- Si falla la conexión, verifica que el servicio `db` esté corriendo

### Paso 5: Poblar la Base de Datos con Datos de Prueba

Ejecuta el seeder para crear datos de ejemplo:

```bash
go run cmd/seed/main.go
```

**Qué crea el seeder:**

**Roles:**

- `admin` (ID: 1)
- `client` (ID: 2)

**Usuarios:**
| Email | Contraseña | Rol |
|-------|------------|-----|
| admin@bsmart.com | admin123 | Admin |
| client@bsmart.com | client123 | Client |
| user1@bsmart.com | password123 | Client |
| user2@bsmart.com | password123 | Client |

**Categorías:** 8 categorías (Electrónica, Ropa, Hogar, etc.)

**Productos:** 20 productos de ejemplo con precios y stock

**Output esperado:**

```
Seeding roles...
✓ Created role: admin
✓ Created role: client
Seeding users...
✓ Created user: admin@bsmart.com
✓ Created user: client@bsmart.com
Seeding categories...
✓ Created category: Electrónica
...
Seeding products...
✓ Created product: Laptop HP Pavilion
...
Seed complete!
```

**Nota**: El seeder es **idempotente**, lo que significa que es seguro ejecutarlo múltiples veces. No creará duplicados.

### Paso 6: Instalar Dependencias de Go

Descarga todas las dependencias del proyecto:

```bash
go mod download
```

### Paso 7: Compilar la Aplicación

Compila el ejecutable de la aplicación:

```bash
go build -o bin/cmd-app ./cmd/app
```

**Verificación**: Deberías ver un nuevo archivo en `bin/cmd-app`

### Paso 8: Ejecutar la Aplicación

Ejecutar directamente con Go:

```bash
go run cmd/app/main.go
```

**Output esperado:**

```
Database pool created successfully
Connected to PostgreSQL successfully
Starting Hub goroutine...
Server starting on :8080
```

**¡La aplicación está corriendo!**

### Paso 9: Verificar la Instalación

Abre una **nueva terminal** y prueba el endpoint de salud:

```bash
curl http://localhost:8080/health
```

**Respuesta esperada:**

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

**Si no tienes `curl`**, abre tu navegador y visita: `http://localhost:8080/health`

### Solución de Problemas Comunes

**Error: "bind: address already in use"**

- Solución: El puerto 8080 ya está en uso. Cambia `PORT=8081` en tu archivo `.env`

**Error: "failed to connect to database"**

- Verifica que los servicios Docker estén corriendo: `docker-compose ps`
- Verifica la variable `DATABASE_URL` en tu `.env`
- Revisa logs de PostgreSQL: `docker-compose logs db`

**Error: "no such table: products"**

- Las migraciones no se ejecutaron. Ejecuta: `docker-compose run --rm migrate up`

**Error: "cannot find package"**

- Las dependencias no están instaladas. Ejecuta: `go mod download`

---

### Características Clave de la Base de Datos

**Índices**:

- `idx_products_name` (GIN): Búsqueda full-text en nombres de productos
- `idx_products_price`: Filtrado rápido por precio
- `idx_products_stock`: Filtrado rápido por stock

**Triggers**:

- `trg_product_history`: Registra automáticamente cambios de precio/stock

**Constraints**:

- Precio y stock deben ser no negativos (constraints CHECK)
- Email debe ser único
- Nombres de categorías deben ser únicos
- Claves foráneas aseguran integridad referencial

---

## WebSockets

La aplicación implementa WebSockets para notificar a los clientes conectados sobre cambios en productos y categorías en tiempo real.

### Implementación

El sistema utiliza el **patrón Hub-Client**:

- **Hub**: Gestor centralizado que mantiene todas las conexiones WebSocket activas
- **Client**: Cada conexión WebSocket se maneja en goroutines independientes para lectura/escritura
- **Broadcasting**: Cuando ocurre un cambio (crear/actualizar/eliminar), el evento se emite automáticamente a todos los clientes conectados

### Eventos Disponibles

Los siguientes eventos se emiten automáticamente cuando se realizan operaciones desde la API:

**Productos:**

- `product:created` - Se creó un nuevo producto
- `product:updated` - Se actualizó un producto (precio, stock, nombre, etc.)
- `product:deleted` - Se eliminó un producto

**Categorías:**

- `category:created` - Se creó una nueva categoría
- `category:updated` - Se actualizó una categoría
- `category:deleted` - Se eliminó una categoría

### Formato de Mensaje

Todos los eventos se envían en formato JSON:

```json
{
  "event": "product:created",
  "data": {
    "id": 1,
    "name": "Laptop HP",
    "price": 899.99,
    "stock": 10,
    ...
  }
}
```

### Probar WebSockets con wscat

**wscat** es una herramienta de línea de comandos para probar conexiones WebSocket.

**Instalación:**

```bash
npm install -g wscat
```

**Conectarse en Local:**

```bash
wscat -c ws://localhost:8080/ws
```

**Conectarse en Producción:**

```bash
wscat -c wss://bsmart-challenge.onrender.com/ws
```

**Una vez conectado**, verás el mensaje `Connected`. Deja la conexión abierta y realiza operaciones en la API (crear/actualizar/eliminar productos o categorías). Los eventos llegarán automáticamente a tu terminal:

```
Connected (press CTRL+C to quit)
< {"event":"product:created","data":{"id":21,"name":"Nuevo Producto","price":50,"stock":100,...}}
< {"event":"product:updated","data":{"id":1,"name":"Laptop HP","price":799.99,"stock":5,...}}
```

---

### Gestión de Base de Datos

```bash
# Crear nueva migración
migrate create -ext sql -dir internal/migrations -seq nombre_migracion

# Verificar versión de migración
docker-compose run --rm migrate version

# Revertir última migración
docker-compose run --rm migrate down 1

# Forzar versión de migración (si está atascada)
docker-compose run --rm migrate force VERSION
```

### Docker

```bash
# Iniciar todos los servicios
docker-compose up -d

# Detener todos los servicios
docker-compose down
```

---

## Consideraciones de Performance

### Optimización de Base de Datos

- **Índices**: Índices estratégicos en campos frecuentemente consultados
- **Connection Pooling**: pgxpool gestiona reutilización de conexiones
- **Prepared Statements**: Queries usan declaraciones parametrizadas
- **Paginación**: Limita datos transferidos por petición

### Escalabilidad de WebSocket

- **Patrón Hub**: Gestión centralizada de conexiones
- **Goroutines**: Cada cliente manejado en goroutine separada
- **Channel Buffering**: Previene bloqueo en clientes lentos
- **Preparado para Redis**: Arquitectura soporta Redis pub/sub para despliegue multi-instancia

---

## Consideraciones de Seguridad

### Implementado

- ✅ Expiración de tokens JWT (24 horas)
- ✅ Hashing de contraseñas con bcrypt
- ✅ Control de acceso basado en roles

---

## Documentos Adicionales

- **[DECISIONES_DE_DISENO.md](./DECISIONES_DE_DISENO.md)**: Documento detallado de decisiones técnicas y arquitectónicas

---

## 👥 Autor

Bruno Malagoli
