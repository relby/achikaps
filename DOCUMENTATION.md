# Документация

### Модели
- Нода (Node)
```json
{
    "ID": uint
    "Type": uint // 1 - Transit, 2 - Production
    "Name": uint
    "Position": {"X": float64, "Y": float64}
    "Radius": float64
    "BuildProgress": float64 // Значение от 0 до 1, если 1 то нода построена
}
```

- Юнит (Unit)
```json
{
    "Type": uint // 1 - Idle, 2 - Production, 3 - Builder, 4 - Transport
    "Node": Node // см. выше
    "Material": Material // только для Transport типа
    "Action": UnitAction // см. ниже
}
```

- Действие юнита (UnitAction)
```json
{
    "Type": // 1 - Moving, 2 - Production, 3 - Building
    "IsStarted": bool
    "Data": any // Данные, зависящие от типа
}

// MovingUnitActionData
{
    "Speed": float64
    "FromNode": Node // От этой ноды юнит начал движение
    "ToNode": Node // К этой ноде движется юнит
    "Progress": float64 // Значение от 0 до 1, обозначающее продвижение по дороге от одной ноде к другой
}

// ProductionUnitActionData 
{
    "Progress": float64
}

// TakeMaterialUnitActionData
{
    "Material": Material
}
```

- Материал (Material)
```json
{
    "ID": uint
    "Type": uint // TODO
    "Node": Node
    "IsReserved": bool
}
```


### Оп коды
- 1. Получение стартого стэйта
  - Ответ:
    ```json
    {
        "Nodes": Map<UserID, Map<NodeID, Node>>
        "Units": Map<UserID, Map<UnitID, Unit>>
        "Materials": Map<UserID, Map<MaterialID, Material>>
    }
    ```
- 2. Строительство ноды
  - Запрос:
    ```json
    {
        "FromNodeID": uint // От какой ноды строить путь до новой
        "Type": uint // Тип ноды, описан ниже
        "Position": {"X": float64, "Y": float64} // Позиция новой ноды
        "Data": any // Допольнительный данные, описаны ниже
    }
    ```
  - Типы нод:
    1. Транзитная нода
    2. Производственная нода
       - Данные, тип `uint` (поле `Data`):
         1) Производит новых юнитов
         2) Производит материал (пока одного типа)
  - Ответа:
    1. Успех: 
    ```json
    {
        "FromNodeID": uint
        "Node": Node
    }
    ```
    2. Ошибка: `{"error": string}`
- 3. Начало выполнение действия юнитом
  - Ответ: Модель UserAction
- 4. Изменение типа юнита
  - Запрос:
  ```json
  {
    "ID": uint
    "Type": uint
  }
  ```
  - Ответ:
    1. Успех: 
    ```json
    {
        "Unit": Unit
    }
    ```
    2. Ошибка: `{"error": string}`