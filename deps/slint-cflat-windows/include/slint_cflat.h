/**
 * @file slint_cflat.h
 * @brief Flat C API bindings for the Slint declarative GUI framework.
 */

#ifndef SLINT_CFLAT_H
#define SLINT_CFLAT_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

/* DLL import/export macros */
#if defined(_WIN32) && !defined(SLINT_CFLAT_STATIC)
  #if defined(SLINT_CFLAT_BUILD_DLL)
    #define SLINT_API __declspec(dllexport)
  #else
    #define SLINT_API __declspec(dllimport)
  #endif
#else
  #define SLINT_API
#endif

#ifdef __cplusplus
extern "C" {
#endif

/* ========================================================================== */
/* 1. Message Identifiers (Windows-like event model)                           */
/* ========================================================================== */

#define SLINT_WM_NULL     0x0000
#define SLINT_WM_CREATE   0x0001
#define SLINT_WM_DESTROY  0x0002
#define SLINT_WM_CLOSE    0x0010
#define SLINT_WM_QUIT     0x0012
#define SLINT_WM_COMMAND  0x0111
#define SLINT_WM_USER     0x0400

/* ========================================================================== */
/* 2. Type Definitions (Must precede function prototypes)                     */
/* ========================================================================== */

/**
 * @brief Standard message packet passed between Slint and the host application.
 */
typedef struct SlintMessage {
    uint32_t  msg;     /**< Message ID. */
    uintptr_t wparam;  /**< Unsigned pointer-sized parameter. */
    intptr_t  lparam;  /**< Signed pointer-sized parameter. */
} SlintMessage;

/**
 * @brief Callback invocation forwarded from a Slint component.
 *
 * The returned strings are owned by SlintContext and remain valid until the
 * next successful call to slint_c_poll_callback() or slint_c_free(). Do not free them.
 * For callbacks declared by an exported global singleton, name is qualified
 * as "GlobalName.callback_name" (for example "AppRouter.open_file").
 * string_arg is NULL when the callback has no string argument (or its first
 * argument is not a string).
 */
typedef struct SlintCallbackEvent {
    const char *name;        /**< Callback name. */
    const char *string_arg;  /**< First string argument, if present. */
} SlintCallbackEvent;

/**
 * @brief Generic function pointer callback signature.
 */
typedef void (*SlintCallback)(void *user_data);

/**
 * @brief Opaque application context handle representing an active Slint instance.
 */
typedef struct SlintContext SlintContext;

typedef enum SlintErrorCode {
    SLINT_ERROR_NONE = 0,
    SLINT_ERROR_INVALID_ARGUMENT = 1,
    SLINT_ERROR_INVALID_UTF8 = 2,
    SLINT_ERROR_COMPILE = 3,
    SLINT_ERROR_NO_COMPONENT = 4,
    SLINT_ERROR_CREATE = 5,
    SLINT_ERROR_CALLBACK = 6,
    SLINT_ERROR_SHOW = 7,
    SLINT_ERROR_EVENT_LOOP = 8,
    SLINT_ERROR_INTERNAL = 9
} SlintErrorCode;

/**
 * @brief Opaque Slint value owned by the C caller.
 */
typedef struct SlintValue SlintValue;

/**
 * @brief Element type for a dynamically created Slint array value.
 *
 * Integer and float elements are both represented as Slint numbers, but
 * appending to an array still requires the matching C value constructor.
 */
typedef enum SlintArrayElementType {
    SLINT_ARRAY_ELEMENT_INTEGER = 1,
    SLINT_ARRAY_ELEMENT_FLOAT = 2,
    SLINT_ARRAY_ELEMENT_STRING = 3,
    SLINT_ARRAY_ELEMENT_IMAGE = 4,
    SLINT_ARRAY_ELEMENT_BOOL = 5
} SlintArrayElementType;

/* ========================================================================== */
/* 3. Lifecycle Management APIs                                               */
/* ========================================================================== */

SLINT_API SlintContext* slint_c_create(const char *code_ptr);
SLINT_API void          slint_c_free(SlintContext *ctx_ptr);
SLINT_API bool          slint_c_is_running(SlintContext *ctx_ptr);
/* The returned error string is thread-local and remains valid until the next
 * error update or slint_c_clear_error() on the calling thread. */
SLINT_API uint32_t      slint_c_get_last_error_code(void);
SLINT_API const char*   slint_c_get_last_error(void);
SLINT_API void          slint_c_clear_error(void);

/**
 * @brief Show the native open-file dialog owned by the Slint application window.
 *
 * The function blocks until the dialog is closed. On success, the selected
 * UTF-8 path is copied to out_path including its NUL terminator. The return
 * value is false when the user cancels or an error occurs; use the last-error
 * API to distinguish those cases.
 */
SLINT_API bool          slint_c_open_file_dialog(
    SlintContext *ctx_ptr,
    char *out_path,
    size_t out_path_capacity
);

/* ========================================================================== */
/* 4. Message Transmission APIs                                               */
/* ========================================================================== */

SLINT_API bool          slint_c_post_message(SlintContext *ctx_ptr, uint32_t msg, uintptr_t wparam, intptr_t lparam);
SLINT_API intptr_t      slint_c_send_message(SlintContext *ctx_ptr, uint32_t msg, uintptr_t wparam, intptr_t lparam);

/* ========================================================================== */
/* 5. Event Loop & Message Polling APIs                                       */
/* ========================================================================== */

SLINT_API bool          slint_c_poll_event(SlintContext *ctx_ptr, SlintMessage *out_msg);
SLINT_API bool          slint_c_dispatch_event(SlintContext *ctx_ptr, const SlintMessage *msg);

/* ========================================================================== */
/* 6. Callback Polling APIs                                                   */
/* ========================================================================== */

SLINT_API bool          slint_c_poll_callback(SlintContext *ctx_ptr, SlintCallbackEvent *event);

/**
 * @brief Set a public Slint property from the C side.
 *
 * value_ptr points to a NUL-terminated UTF-8 string for string properties, a
 * bool for bool properties, a double for numeric properties, or a SlintValue
 * containing an image for image properties. Values are copied before the
 * update is queued on the Slint event loop.
 */
SLINT_API bool          slint_c_set_property(SlintContext *ctx_ptr, const char *name, const void *value_ptr);

/**
 * @brief Set a property on an exported global singleton.
 *
 * value_ptr uses the same representation as slint_c_set_property().
 */
SLINT_API bool          slint_c_set_global_property(
    SlintContext *ctx_ptr,
    const char *global_name,
    const char *property_name,
    const void *value_ptr
);

/**
 * @brief Get a property from an exported global singleton.
 *
 * out_value points to bool or double storage for those property types. For a
 * string property it must be a const char **; the returned string is owned by
 * SlintContext and remains valid until the next getter call or slint_c_free().
 */
SLINT_API bool          slint_c_get_global_property(
    SlintContext *ctx_ptr,
    const char *global_name,
    const char *property_name,
    void *out_value
);

/* ========================================================================== */
/* 7. Slint Value Handles                                                     */
/* ========================================================================== */

/**
 * @brief Load an image from a UTF-8 file path.
 *
 * Returns NULL on failure; use the last-error API for details.
 */
SLINT_API SlintValue*   slint_c_image_load_from_file(const char *filename);

SLINT_API SlintValue*   slint_c_value_new_string(const char *value);
SLINT_API SlintValue*   slint_c_value_new_bool(bool value);
SLINT_API SlintValue*   slint_c_value_new_number(double value);
SLINT_API SlintValue*   slint_c_value_new_integer(int64_t value);
SLINT_API SlintValue*   slint_c_value_new_float(float value);
/**
 * @brief Create an empty Slint model value for the selected element type.
 *
 * Returns NULL for an unsupported element type.
 */
SLINT_API SlintValue*   slint_c_value_new_array(uint32_t element_type);
/**
 * @brief Append a copy of an element to an array value.
 *
 * element_type must match the array's declared type and the actual SlintValue
 * type. Returns false if either check fails or array is not an array value.
 */
SLINT_API bool          slint_c_value_array_append(
    SlintValue *array,
    uint32_t element_type,
    const SlintValue *element
);
SLINT_API void          slint_c_value_free(SlintValue *value);

SLINT_API bool          slint_c_set_property_value(
    SlintContext *ctx_ptr,
    const char *name,
    const SlintValue *value
);

SLINT_API bool          slint_c_set_global_property_value(
    SlintContext *ctx_ptr,
    const char *global_name,
    const char *property_name,
    const SlintValue *value
);

#ifdef __cplusplus
}
#endif

#endif /* SLINT_CFLAT_H */
