# Decisiones de Diseño - Bsmart Backend Challenge

Este documento describe las principales decisiones de diseño tomadas durante el desarrollo del proyecto y su justificación técnica.

---

## 1. Arquitectura Handler → DB → PostgreSQL

**Decisión**: Utilizar una arquitectura simple de Handler → DB → Base de Datos sin capas de servicio adicionales.

**Justificación**:

- Desarrollo más rápido para el alcance del challenge
- Código más fácil de entender y mantener
- Posibilidad de refactorizar añadiendo capa de servicio en el futuro si es necesario

**Trade-offs**: Menor abstracción significa que los handlers tienen más responsabilidades, pero para este tamaño de proyecto es aceptable.

---

## 2. Triggers de PostgreSQL para Historial de Productos

**Decisión**: Usar triggers de base de datos para registrar automáticamente cambios en productos.

**Implementación**:

```sql
CREATE TRIGGER trg_product_history
AFTER UPDATE ON products
FOR EACH ROW
EXECUTE PROCEDURE fn_product_history();
```

**Justificación**:

- **Confiabilidad**: Imposible olvidar registrar cambios
- **Performance**: Operación del lado de la base de datos, sin viajes adicionales
- **Atomicidad**: El historial se registra en la misma transacción que la actualización
- **Consistencia**: Funciona sin importar cómo se actualicen los productos

---

## 3. Patrón Hub-Client para WebSockets

**Decisión**: Usar gorilla/websocket con un Hub que gestiona todos los clientes conectados.

**Arquitectura**:

- **Hub**: Gestor central con canales para register/unregister/broadcast
- **Client**: Conexión individual con goroutines ReadPump y WritePump
- **Event Broadcasting**: Los handlers emiten eventos después de operaciones exitosas en BD

**Justificación**:

- **Escalable**: El patrón Hub es probado para gestionar muchas conexiones
- **Thread-safe**: Los canales de Go manejan concurrencia naturalmente
- **Desacoplado**: Los handlers no gestionan conexiones WebSocket directamente
- **Preparado para Redis**: Fácil de extender con Redis pub/sub para despliegue multi-instancia

---

## 4. JWT con Autorización Basada en Roles

**Decisión**: Tokens JWT con claims de rol embebidos, middleware para protección de rutas.

**Flujo**:

```
1. Usuario inicia sesión → JWT generado con {user_id, email, role}
2. Cliente envía token en Authorization: Bearer <token>
3. Middleware valida token, extrae claims, los almacena en contexto de Gin
4. Middleware de rol verifica si el usuario tiene el rol requerido
5. Handler accede a la información del usuario desde el contexto
```

**Justificación**:

- **Stateless**: No requiere almacenamiento de sesiones en el servidor
- **Escalable**: Funciona en múltiples instancias de servidor
- **Flexible**: Fácil de agregar más claims o permisos
- **Estándar**: Enfoque estándar de la industria

**Medidas de seguridad**:

- Hashing de contraseñas con bcrypt (factor de costo 10)
- Expiración de tokens (24 horas)
- Firma HMAC-SHA256
- Mensajes de error apropiados (no filtran información)

---

## 5. Búsqueda Full-Text con Índice GIN de PostgreSQL

**Decisión**: Usar búsqueda full-text nativa de PostgreSQL con índice GIN.

**Implementación**:

```sql
CREATE INDEX idx_products_name ON products
USING gin (to_tsvector('simple', name));
```

**Query**:

```sql
WHERE to_tsvector('simple', name) @@ plainto_tsquery('simple', $1)
```

**Justificación**:

- **Performance**: Los índices GIN son extremadamente rápidos para búsqueda de texto
- **Nativo**: No requiere motor de búsqueda externo (Elasticsearch, etc.)
- **Simple**: Menos infraestructura que gestionar

**Trade-off**: Para búsqueda más avanzada (sinónimos, fuzzy matching), Elasticsearch sería mejor.

---

## 6. Formato de Respuesta API Consistente

**Decisión**: Estandarizar todas las respuestas de la API con estructura de éxito/error.

**Formato**:

```json
{
  "success": true,
  "data": { ... }
}

{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Producto no encontrado"
  }
}
```

**Justificación**:

- **Amigable para el cliente**: Fácil de parsear y manejar
- **Consistente**: Todos los endpoints siguen el mismo patrón
- **Informativo**: Los códigos de error ayudan con el debugging
- **Profesional**: Práctica estándar de APIs REST

---

## 7. Docker Compose para Desarrollo

**Decisión**: Usar Docker Compose para el entorno de desarrollo local.

**Servicios**:

- PostgreSQL 15
- App (aplicación Go y Seeder)
- Migrate (ejecutor de migraciones)

**Justificación**:

- **Consistencia**: Mismo entorno para todos los desarrolladores
- **Aislamiento**: Sin conflictos con bases de datos instaladas en el sistema
- **Conveniencia**: Un comando para iniciar todo
- **Similar a producción**: Entorno similar al de despliegue

---

## 8. Gestión de Schema Basada en Migraciones

**Decisión**: Usar golang-migrate para versionado del schema de base de datos.

**Justificación**:

- **Control de versiones**: Los cambios de schema se rastrean en Git
- **Reproducible**: Mismo schema en todos los entornos
- **Rollback**: Se pueden deshacer migraciones si es necesario
- **Amigable para equipos**: Múltiples desarrolladores pueden trabajar en el schema

---

## 9. CLI de Seeder Separado

**Decisión**: Crear ejecutable de seeder dedicado en lugar de endpoint de API.

**Justificación**:

- **Seguridad**: No exponer seeding vía HTTP
- **Flexibilidad**: Se puede ejecutar independientemente
- **Idempotente**: Seguro ejecutar múltiples veces
- **Desarrollo**: Fácil resetear datos de prueba

### 10. Índices Estratégicos en Campos Frecuentemente Consultados

**Decisión**: Crear índices en campos como `price`, `stock` y `name` de productos.

**Justificación**:

- **Filtrado rápido**: Queries por precio y stock son comunes
- **Búsqueda optimizada**: Índice GIN para full-text search en nombres
- **Balance**: No sobre-indexar (cada índice tiene costo en writes)

---

### 11. Validación de Datos en la Capa de Handler

**Decisión**: Validar datos de entrada en los handlers antes de procesarlos.

**Justificación**:

- **Seguridad**: Prevenir datos inválidos o maliciosos
- **Experiencia del usuario**: Errores claros y tempranos
- **Integridad**: Proteger la base de datos de datos inconsistentes
- **Constraints DB**: Combinado con constraints de base de datos (CHECK, NOT NULL)

---

## Resumen

Las decisiones tomadas priorizan:

- **Simplicidad** sobre complejidad innecesaria
- **Performance** con índices y connection pooling
- **Seguridad** con JWT, bcrypt y validaciones
- **Mantenibilidad** con código claro y explícito
- **Escalabilidad** con arquitectura preparada para crecer

Todas las decisiones consideran el balance entre velocidad de desarrollo para el challenge y la posibilidad de evolucionar el código hacia un sistema de producción robusto.
