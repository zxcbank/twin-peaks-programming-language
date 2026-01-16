package bytecode

const (
	OpInvalid    byte = iota // Недействительная операция
	OpConst                  // Загрузить константу
	OpLoad                   // Загрузить переменную
	OpStore                  // Сохранить в переменную
	OpPop                    // Удалить верхнее значение стека
	OpAdd                    // Сложение
	OpSub                    // Вычитание
	OpMul                    // Умножение
	OpDiv                    // Деление
	OpMod                    // Остаток от деления
	OpNeg                    // Унарный минус
	OpEq                     // Равенство
	OpNeq                    // Неравенство
	OpLt                     // Меньше
	OpLe                     // Меньше или равно
	OpGt                     // Больше
	OpGe                     // Больше или равно
	OpAnd                    // Логическое И
	OpOr                     // Логическое ИЛИ
	OpNot                    // Логическое НЕ
	OpJmp                    // Безусловный переход
	OpJmpIfFalse             // Переход если false
	OpCall                   // Вызов функции
	OpReturn                 // Возврат из функции
	OpReturnVoid             // Возврат из функции без значения
	OpPrint                  // Вывод
	OpSqrt
	OpHalt // Остановка

	OpArrayAlloc
	OpArrayLoad
	OpArrayStore
)
