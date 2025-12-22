cmake_minimum_required(VERSION 3.22)

project(TelecomBaseClient LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTOUIC ON)
set(CMAKE_AUTORCC ON)

find_package(Qt6 REQUIRED COMPONENTS Widgets Network)

add_executable(TelecomBaseClient
    src/main.cpp
    src/apiClient.cpp
    src/authDialog.cpp
    src/deviceDialog.cpp
    src/vendorsDialog.cpp
    src/modelsDialog.cpp
    src/locationsDialog.cpp
    src/usersDialog.cpp
    src/mainWindow.cpp
    src/themeManager.cpp
    resources/resources.qrc
    include/apiClient.h
    include/models.h
    include/authDialog.h
    include/deviceDialog.h
    include/vendorsDialog.h
    include/modelsDialog.h
    include/locationsDialog.h
    include/usersDialog.h
    include/mainWindow.h
    include/themeManager.h
)

target_include_directories(TelecomBaseClient PRIVATE include)

target_link_libraries(TelecomBaseClient PRIVATE Qt6::Widgets Qt6::Network)
