package bytecode

const (
	// Константы инструкций
	OP_CONST        byte = iota // Загрузить константу
	OP_LOAD                     // Загрузить переменную
	OP_STORE                    // Сохранить в переменную
	OP_POP                      // Удалить верхнее значение стека
	OP_ADD                      // Сложение
	OP_SUB                      // Вычитание
	OP_MUL                      // Умножение
	OP_DIV                      // Деление
	OP_MOD                      // Остаток от деления
	OP_NEG                      // Унарный минус
	OP_EQ                       // Равенство
	OP_NEQ                      // Неравенство
	OP_LT                       // Меньше
	OP_LE                       // Меньше или равно
	OP_GT                       // Больше
	OP_GE                       // Больше или равно
	OP_AND                      // Логическое И
	OP_OR                       // Логическое ИЛИ
	OP_NOT                      // Логическое НЕ
	OP_JMP                      // Безусловный переход
	OP_JMP_IF_FALSE             // Переход если false
	OP_CALL                     // Вызов функции
	OP_RETURN                   // Возврат из функции
	OP_PRINT                    // Вывод
	OP_SQRT
	OP_HALT // Остановка

	OP_RETURN_VOID

	OP_ARRAY_ALLOC
	OP_ARRAY_LOAD
	OP_ARRAY_STORE
)

// Типы значений в байт-коде
const (
	TypeInt = iota
	TypeFloat
	TypeString
	TypeBool
	TypeVoid
)
