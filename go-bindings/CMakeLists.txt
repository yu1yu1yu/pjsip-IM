cmake_minimum_required(VERSION 3.28)
project(pjgo_binding LANGUAGES C CXX)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(BUILD_TESTING OFF CACHE BOOL "" FORCE)
set(BUILD_SHARED_LIBS OFF CACHE BOOL "" FORCE)
set(PJ_SKIP_EXPERIMENTAL_NOTICE ON CACHE BOOL "" FORCE)
set(PJLIB_WITH_SSL "" CACHE STRING "" FORCE)
set(PJMEDIA_WITH_VIDEO OFF CACHE BOOL "" FORCE)
set(PJMEDIA_WITH_AUDIODEV OFF CACHE BOOL "" FORCE)
set(PJMEDIA_WITH_SRTP OFF CACHE BOOL "" FORCE)
add_subdirectory(.. pjproject EXCLUDE_FROM_ALL)
add_library(pjgo SHARED native/bridge.c)
target_compile_definitions(pjgo PRIVATE PJGO_BUILD)
target_link_libraries(pjgo PRIVATE pjsua-lib)
foreach(part pjlib pjlib-util pjnath pjmedia pjsip)
  target_include_directories(pjgo PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/../${part}/include"
    "${CMAKE_CURRENT_BINARY_DIR}/pjproject/${part}/include")
endforeach()
set_target_properties(pjgo PROPERTIES
  RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
  LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin"
  ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
if(MINGW)
  target_link_options(pjgo PRIVATE -static-libgcc -static-libstdc++)
endif()
