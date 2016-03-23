package io.iron.lambda;

import java.io.*;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.*;
import java.util.stream.*;

import com.google.gson.JsonSyntaxException;
import io.github.lukehutch.fastclasspathscanner.FastClasspathScanner;

public class Launcher {
    public static void main(String[] args) {
        String handler = args[0];
        String payload = "";
        try {
            String file = System.getenv("PAYLOAD_FILE");
            if (file != null) {
                payload = new String(Files.readAllBytes(Paths.get(file)));
            }
        } catch (IOException ioe) {
            // Should probably log this somewhere useful but not in the output.
        }

        try {
            String[] packageMethod = validateInputParamsAndGetPackageMethod(handler, payload);
            Launcher ll = new Launcher();
            ll.launchMethod(packageMethod, payload);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static String[] validateInputParamsAndGetPackageMethod(String handler, String payload) {
        if (handler == null) {
            System.err.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        if (payload == null) {
            System.err.println("Payload should not be null. There is a problem in the Launcher.");
            System.exit(1);
        }
        String[] package_function = handler.split("::");
        if (package_function[0] == null || package_function[0].equals("") || package_function[1] == null || package_function[1].equals("")) {
            System.err.println("Handler is not specified, please specify handler function (for example: 'example.Hello::myHandler')");
            System.exit(1);
        }
        return package_function;
    }

    private Object callWithMaybeContext(Method lambdaMethod, Object instance, boolean contextRequired, Object arg, AWSContext ctx) throws IllegalAccessException, java.lang.reflect.InvocationTargetException {
      if (contextRequired) {
        return lambdaMethod.invoke(instance, arg, ctx);
      }
      return lambdaMethod.invoke(instance, arg);
    }

    private void runMethod(Method lambdaMethod, String payload, Class cls, String handlerName) throws Exception {
        Object result = null;
        boolean processed = false;
        AWSContext aws_ctx = new AWSContext();

        Object lambdaInstance = cls.newInstance();

        String packageName = lambdaInstance.getClass().getPackage().getName();

        List<String> classNames = new FastClasspathScanner(packageName).scan().getNamesOfAllClasses();
        String packageNamePrefix = packageName + ".";

        Class[] parameterTypes = lambdaMethod.getParameterTypes();

        boolean contextRequired = parameterTypes.length > 0 &&
                                  parameterTypes[parameterTypes.length-1].getTypeName().equals("com.amazonaws.services.lambda.runtime.Context");

        if (parameterTypes.length == 0) {
          lambdaMethod.invoke(lambdaInstance);
          processed = true;
        } else if (parameterTypes.length == 1 && parameterTypes[0].getTypeName().equals("com.amazonaws.services.lambda.runtime.Context")) {
          lambdaMethod.invoke(lambdaInstance, aws_ctx);
          processed = true;
        // IOStreams
        } else if (checkIfLambdaMethodRequiredIOStreams(parameterTypes)) {
            ByteArrayOutputStream out = new ByteArrayOutputStream();
            lambdaMethod.invoke(lambdaInstance, new ByteArrayInputStream(payload.getBytes(StandardCharsets.UTF_8)), out, aws_ctx);
            result = new String(out.toByteArray(), "UTF-8");
            processed = true;
        // POJO, map, list
        } else if (checkIfLambdaMethodRequiredPOJO(parameterTypes, classNames, packageNamePrefix) || checkIfLambdaMethodRequiredMap(parameterTypes) || checkIfLambdaMethodRequiredList(parameterTypes)) {
            result = ClassTypeHelper.gson.toJson(
                         callWithMaybeContext(lambdaMethod, lambdaInstance,
                                              contextRequired,
                                              ClassTypeHelper.gson.fromJson(
                                                  payload,
                                                  parameterTypes[0]),
                                              aws_ctx));
            processed = true;
        // int, string, bool
        } else if (checkIfLambdaMethodRequiredPrimitiveType(parameterTypes)) {
            if (parameterTypes[0].getName().equals("java.lang.String")) {
                String p = ClassTypeHelper.gson.fromJson(payload, String.class);
                result = callWithMaybeContext(lambdaMethod, lambdaInstance, contextRequired, p, aws_ctx);
                processed = true;
            } else if (parameterTypes[0].toString().equals("int") || Objects.equals(parameterTypes[0].toString(), "Integer")) {
                int i = ClassTypeHelper.gson.fromJson(payload, int.class);
                result = callWithMaybeContext(lambdaMethod, lambdaInstance, contextRequired, i, aws_ctx);
                processed = true;
            } else if (parameterTypes[0].toString().equals("boolean") || Objects.equals(parameterTypes[0].toString(), "Boolean")) {
                boolean b = ClassTypeHelper.gson.fromJson(payload, boolean.class);
                result = callWithMaybeContext(lambdaMethod, lambdaInstance, contextRequired, b, aws_ctx);
                processed = true;
            }
        }
        if (!processed) {
            System.err.println(String.format("Handler %s with simple, POJO, or IO(input/output) types not found", handlerName));
        }
    }

    private void launchMethod(String[] packageHandler, String payload) throws Exception {
        String packageClass = packageHandler[0];
        String handlerName = packageHandler[1];

        Class cls = Class.forName(packageClass); //get class in package
        Method[] declaredMethods = cls.getDeclaredMethods();

        // get only methods matching handler and accepting 0 <= N <=
        // 3 arguments.
        // If the first argument is a valid type, the payload is passed to it
        // (if the payload can be deserialized to that type).
        // If the first argument is a Context, no payload is passed.
        // If their are 0 arguments, the function is just called.
        // The 3 argument form is allowed in case of using streams.
        Stream<Method> matches = Arrays.stream(declaredMethods).filter(m -> {
          if (!m.getName().equals(handlerName)) {
            return false;
          }

          Class[] pt = m.getParameterTypes();
          return checkIfLambdaMethodRequiredIOStreams(pt) || pt.length <= 2;
        });
        Method[] matchedArray = matches.toArray(Method[]::new);

        if (matchedArray.length == 0) {
            System.err.println(String.format("Method %s not found", handlerName));
            System.exit(1);
        }

        // sort according to rule:
        //  1. Select the method with the largest number of parameters.
        //
        //  2. If two or more methods have the same number of parameters, AWS
        //     Lambda selects the method that has the Context as the last
        //     parameter.
        //
        // If none or all of these methods have the Context parameter, then the
        // behavior is undefined.
        Arrays.sort(matchedArray, (a, b) -> {
          Class[] parameterTypesA = a.getParameterTypes();
          Class[] parameterTypesB = b.getParameterTypes();
          // If number of params is the same, the one that has Context as the
          // last param should be first, so it should be 'less than' the other.
          if (parameterTypesA.length == parameterTypesB.length) {
            if (parameterTypesA[parameterTypesA.length - 1].getTypeName().equals("com.amazonaws.services.lambda.runtime.Context")) {
              return -1;
            }
            // Don't bother checking the second one, since it is undefined in
            // case of both being true or false.
            return 1;
          }

          // If A accepts less parameters, it should come later in the list, so
          // it is 'greater than' B.
          if (parameterTypesA.length < parameterTypesB.length) {
            return 1;
          }

          return -1;
        });

        Class returnType = matchedArray[0].getReturnType();
        if (!returnType.getTypeName().equals("void")) {
            System.err.println(String.format("Handler can only have 'void' return type. Found '%s'.", returnType.getTypeName()));
            System.exit(1);
        }
        runMethod(matchedArray[0], payload, cls, handlerName);
    }

    public boolean checkIfLambdaMethodRequiredIOStreams(Class[] parameterTypes) {
        return parameterTypes.length >= 2 && parameterTypes.length <= 3 &&
               parameterTypes[0].getTypeName().equals("java.io.InputStream") &&
               parameterTypes[1].getTypeName().equals("java.io.OutputStream");
    }

    public boolean checkIfLambdaMethodRequiredPOJO(Class[] parameterTypes, List<String> classNames, String packageNamePrefix) {
        boolean classExist = false;
        for (String className : classNames) {
            if (className.startsWith(packageNamePrefix) &&
                className.toLowerCase().contains(parameterTypes[0].getTypeName().toLowerCase())) {
                classExist = true;
                break;
            }
        }
        return parameterTypes.length > 0 && parameterTypes.length <= 2 && classExist;
    }

    private boolean checkIfLambdaMethodRequiredMap(Class[] parameterTypes) {
        return parameterTypes.length > 0 && parameterTypes.length <= 2 &&
               (parameterTypes[0].getTypeName().equals("java.util.Map") ||
                parameterTypes[0].getTypeName().equals("java.util.HashMap"));
    }

    private boolean checkIfLambdaMethodRequiredList(Class[] parameterTypes) {
        return parameterTypes.length > 0 && parameterTypes.length <= 2 &&
               (parameterTypes[0].getTypeName().equals("java.util.ArrayList") ||
                parameterTypes[0].getTypeName().equals("java.util.List"));
    }
    private boolean checkIfLambdaMethodRequiredPrimitiveType(Class[] parameterTypes) {
        return parameterTypes.length > 0 && parameterTypes.length <= 2 &&
               ClassTypeHelper.isSimpleType(parameterTypes[0]);
    }
}
