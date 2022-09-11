# Ejercicio 5

El flujo general de este ejercicio es el siguiente:
El servidor comienza a esperar conexiones de clientes. Cuando un cliente logra conectarse el servidor quedará esperando por el mensaje del cliente. El cliente por su lado arma un array de bytes (siguiendo un protocolo explicado mas adelante) con los datos del jugador del cual se quiere averiguar si es ganador o no. Una vez armado este array, se lo envia al servidor y se queda esperando por la respuesta del servidor. Luego, el servidor lee el mensaje recibido por el cliente, arma un objecto Contestant, llama a la funcion is_winner con dicho objeto y devuelve un byte con la respuesta. 'W' si el Contestant es ganador y 'L' si no lo es. Finalmente el cliente lee este byte e imprime el mensaje correspondiente, mientras el servidor vuelve a esperar por una nueva conexión.

El protocolo previamente mencionado para el armado del mensaje con los datos del jugador es el siguiente:
Primero se construye el buffer, agregando cada campo del jugador, con 1 byte para el largo de dicho campo precediendolo. Luego de armar el buffer se envia el largo total de este en 2 bytes seguido del buffer en si. Entonces, un mensaje completo al server quedaría asi:
(33 0) (4)Ivan(7)Soriano(8)41824203(10)1999-04-11. 
(Se agrego un espacio al principio entre el 33 que es el largo del mensaje total y 4 que es el largo del primer campo para mas claridad, pero en el buffer real este espacio no existe)
