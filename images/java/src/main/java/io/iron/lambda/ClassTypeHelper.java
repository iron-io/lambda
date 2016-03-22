package io.iron.lambda;

import com.google.gson.Gson;

import java.util.HashSet;
import java.util.Set;

class ClassTypeHelper {
    public static final Gson gson = new Gson();
    private static final Set<Class<?>> CLASS_TYPES = getSimpleTypes();

    private ClassTypeHelper() {}

    public static boolean isSimpleType(Class<?> classType) {
        return CLASS_TYPES.contains(classType);
    }

    private static Set<Class<?>> getSimpleTypes() {
        Set<Class<?>> ret = new HashSet<>();
        ret.add(String.class);
        ret.add(Integer.class);
        ret.add(int.class);
        ret.add(Boolean.class);
        ret.add(boolean.class);
        return ret;
    }

}
