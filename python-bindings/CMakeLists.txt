cmake_minimum_required(VERSION 3.28)
project(pjsua2_python_binding LANGUAGES C CXX)

find_package(Python3 REQUIRED COMPONENTS Interpreter Development.Module)
find_package(SWIG 4.0 REQUIRED)
set(PJ_SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/.." CACHE PATH "PJPROJECT source")
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(BUILD_TESTING OFF CACHE BOOL "" FORCE)
set(BUILD_SHARED_LIBS OFF CACHE BOOL "" FORCE)
add_subdirectory("${PJ_SOURCE_DIR}" "${CMAKE_CURRENT_BINARY_DIR}/pjproject" EXCLUDE_FROM_ALL)

set(generated "${CMAKE_CURRENT_BINARY_DIR}/generated")
file(MAKE_DIRECTORY "${generated}")
set(swig_includes)
foreach(part pjlib pjlib-util pjmedia pjsip pjnath)
  list(APPEND swig_includes "-I${PJ_SOURCE_DIR}/${part}/include")
endforeach()
file(GLOB_RECURSE binding_headers CONFIGURE_DEPENDS "${PJ_SOURCE_DIR}/pjsip/include/pjsua2/*.hpp")
add_custom_command(
  OUTPUT "${generated}/pjsua2_wrap.cpp" "${generated}/pjsua2.py"
  COMMAND "${SWIG_EXECUTABLE}" ${swig_includes} -c++ -python -threads -doxygen -w312
          -outdir "${generated}" -o "${generated}/pjsua2_wrap.cpp"
          "${PJ_SOURCE_DIR}/pjsip-apps/src/swig/pjsua2.i"
  DEPENDS "${PJ_SOURCE_DIR}/pjsip-apps/src/swig/pjsua2.i"
          "${PJ_SOURCE_DIR}/pjsip-apps/src/swig/symbols.i" ${binding_headers}
  VERBATIM)
Python3_add_library(_pjsua2 MODULE WITH_SOABI "${generated}/pjsua2_wrap.cpp")
target_compile_features(_pjsua2 PRIVATE cxx_std_11)
target_link_libraries(_pjsua2 PRIVATE pjsua2)
foreach(part pjlib pjlib-util pjnath pjmedia pjsip pjsua-lib pjsua2)
  target_include_directories(_pjsua2 PRIVATE "$<TARGET_PROPERTY:${part},INTERFACE_INCLUDE_DIRECTORIES>")
  target_compile_definitions(_pjsua2 PRIVATE "$<TARGET_PROPERTY:${part},INTERFACE_COMPILE_DEFINITIONS>")
endforeach()
execute_process(COMMAND "${Python3_EXECUTABLE}" -c
  "import sysconfig; print(sysconfig.get_path('platlib'))"
  OUTPUT_VARIABLE python_site OUTPUT_STRIP_TRAILING_WHITESPACE COMMAND_ERROR_IS_FATAL ANY)
install(TARGETS _pjsua2 LIBRARY DESTINATION "${python_site}" COMPONENT PythonBindings)
install(FILES "${generated}/pjsua2.py" DESTINATION "${python_site}" COMPONENT PythonBindings)
